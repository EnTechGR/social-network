package websocket

import (
	"encoding/json"
	"log"
	"social-network/models" // Added import
	"strings"
	"sync"
	"time"
)

// Hub maintains the set of active clients and handles message broadcasting.
// It acts as the central router for all WebSocket traffic.
type Hub struct {
	// Registered clients mapped by user ID.
	// A user can have multiple active WebSocket connections (e.g. sidebar + chat drawer + another tab).
	Clients map[string]map[*Client]struct{} // Key: UserID, Value: set of active client connections

	// Mutex to protect the Clients map from concurrent read/write access.
	// RWMutex allows multiple readers simultaneously but requires exclusive lock for writing (register/unregister).
	mu sync.RWMutex

	// Inbound messages channel used to distribute messages from clients to all other clients.
	// Messages are typically processed by the Hub's Run loop before being sent to connections.
	Broadcast chan []byte

	// Register channel handles requests from new WebSocket clients attempting to join the Hub.
	Register chan *Client

	// Unregister channel handles requests from clients closing their connection (e.g., disconnection or error).
	Unregister chan *Client
}

// NewHub creates a new Hub instance, initializing its internal data structures.
//
// All channels are initialized:
// - Clients map is initialized empty.
// - Broadcast channel is buffered to prevent senders from blocking if the Hub's Run loop is briefly delayed.
// - Register and Unregister channels are unbuffered, ensuring synchronous communication for client lifecycle management.
func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]map[*Client]struct{}),
		// Use a moderately sized buffer for the broadcast channel (256) to handle bursts of messages.
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// BroadcastToUser sends a raw message payload to a single, specific user's active connection.
// This function is thread-safe as it uses a Read Lock (RLock) to access the Clients map.
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	clientSet, exists := h.Clients[userID]
	if !exists || len(clientSet) == 0 {
		h.mu.RUnlock()
		return
	}

	clients := make([]*Client, 0, len(clientSet))
	for client := range clientSet {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
			// Message sent successfully to the client's outgoing buffer.
		default:
			// Slow or dead connection. Unregister this specific client connection.
			log.Printf("Client buffer full for user %s. Attempting graceful shutdown.", userID)
			h.Unregister <- client
		}
	}
}

// Run is the main event loop for the Hub.
//
// It continuously listens on the hub's three primary channels (Register, Unregister, Broadcast)
// and handles client lifecycle management and message routing within a single goroutine.
// This design centralizes state changes, ensuring thread safety for the Clients map.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// A new client is requesting to join.
			h.registerClient(client)

		case client := <-h.Unregister:
			// A client is disconnecting or being marked for removal (e.g., slow reader).
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			// A message needs to be routed to all connected clients.
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub's Clients map.
//
// It handles authentication, concurrent connection management (replacing old sessions),
// and broadcasts presence status updates.
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()

	clientSet, exists := h.Clients[client.UserID]
	if !exists {
		clientSet = make(map[*Client]struct{})
		h.Clients[client.UserID] = clientSet
	}

	wasOffline := len(clientSet) == 0
	clientSet[client] = struct{}{}
	userCount := len(h.Clients)
	connectionCount := 0
	for _, connections := range h.Clients {
		connectionCount += len(connections)
	}

	h.mu.Unlock()

	log.Printf("User %s connected. Online users: %d, connections: %d", client.UserID, userCount, connectionCount)

	// Only broadcast online status for the first active connection for that user.
	if wasOffline {
		h.broadcastOnlineStatus(client.UserID, client.Nickname, true)
	}

	// Send list of online users to this newly connected client.
	h.sendOnlineUsersList(client)
}

// unregisterClient removes a client from the hub's Clients map and cleans up its resources.
//
// This operation is critical for maintaining accurate user presence and preventing resource leaks.
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()

	clientSet, exists := h.Clients[client.UserID]
	if !exists {
		h.mu.Unlock()
		return
	}

	if _, registered := clientSet[client]; !registered {
		h.mu.Unlock()
		return
	}

	delete(clientSet, client)
	close(client.Send)

	becameOffline := false
	if len(clientSet) == 0 {
		delete(h.Clients, client.UserID)
		becameOffline = true
	}

	userCount := len(h.Clients)
	connectionCount := 0
	for _, connections := range h.Clients {
		connectionCount += len(connections)
	}

	h.mu.Unlock()

	log.Printf("User %s disconnected. Online users: %d, connections: %d", client.UserID, userCount, connectionCount)

	// Only broadcast offline when the last active connection for that user closed.
	if becameOffline {
		h.broadcastOnlineStatus(client.UserID, client.Nickname, false)
	}
}

// broadcastMessage routes a raw message payload to the outgoing channel of EVERY connected client.
//
// This is typically used for general, non-targeted messages (e.g., global announcements)
// or for real-time presence updates (handled by auxiliary functions).
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0)
	for _, clientSet := range h.Clients {
		for client := range clientSet {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
			// Message successfully buffered for the client's writer goroutine.
		default:
			log.Printf("Skipping message for user %s (buffer full)", client.UserID)
		}
	}
}

// SendChatMessage formats a structured message and broadcasts it to both the
// receiver and the sender's active WebSocket connections.
//
// This is the primary function for routing new chat messages (text or image)
// after they have been successfully persisted in the database.
func (h *Hub) SendChatMessage(receiverID string, msgData models.MessageWithUser) {
	// 1. Prepare WebSocket message structure.
	msgType := MessageTypeChat // Assume standard chat message type

	// NOTE: The conditional logic for checking msgData.Image is currently commented out,
	// but the structure models.MessageWithUser is designed to handle both types.
	// The client-side application will typically infer the type by checking if the 'Image' field is present.

	wsMsg := WebSocketMessage{
		Type: msgType,
		// Data contains the full message context needed by the client (sender name, IDs, content, image metadata).
		Data:      msgData,
		Timestamp: time.Now(),
	}

	// Serialize the structured message into a JSON byte array for transmission.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message: %v", err)
		return
	}

	// 2. Targeted delivery. Use BroadcastToUser for safe, synchronous routing.

	// Broadcast to Receiver: Ensure the recipient sees the new message in real-time.
	h.BroadcastToUser(receiverID, message)

	// Send back to Sender: Crucial for multi-device support. The sender's other
	// active connections (e.g., phone, desktop) need to be updated with the sent message.
	h.BroadcastToUser(msgData.SenderID, message)
}

// SendGroupChatMessage broadcasts a group chat message to all connected group members.
func (h *Hub) SendGroupChatMessage(memberIDs []string, msgData models.GroupMessage) {
	wsMsg := WebSocketMessage{
		Type:      MessageTypeGroupChat,
		Data:      msgData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal group chat message: %v", err)
		return
	}

	seen := make(map[string]struct{}, len(memberIDs))
	for _, memberID := range memberIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}
		if _, exists := seen[memberID]; exists {
			continue
		}
		seen[memberID] = struct{}{}
		h.BroadcastToUser(memberID, message)
	}
}

// SendMessageDeleteNotification creates and sends a notification signal to both the
// sender and receiver indicating that a message has been deleted.
func (h *Hub) SendMessageDeleteNotification(userID, messageID, senderID string) {
	// 1. Prepare data payload containing necessary deletion identifiers.
	deleteData := map[string]string{
		"message_id": messageID,
		"sender_id":  senderID, // Needed by the receiver to locate the conversation
	}

	// 2. Wrap the payload in the WebSocket message envelope.
	wsMsg := WebSocketMessage{
		Type:      MessageTypeMessageDelete,
		Data:      deleteData, // The raw data payload
		Timestamp: time.Now(),
	}

	// Serialize the notification into a JSON byte array.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal delete notification: %v", err)
		return
	}

	// 3. Notify the receiver (the user who received the message).
	h.BroadcastToUser(userID, message)

	// 4. Notify the sender (the user who performed the deletion),
	// to update their own UI on any other connected devices.
	h.BroadcastToUser(senderID, message)
}

// broadcastOnlineStatus sends a real-time presence update (online/offline)
// to all currently connected clients, excluding the user whose status changed.
func (h *Hub) broadcastOnlineStatus(userID, username string, isOnline bool) {
	// 1. Prepare the payload data.
	statusData := OnlineStatusData{
		UserID:   userID,
		Username: username,
		IsOnline: isOnline,
	}

	// 2. Wrap the payload in the WebSocket message envelope.
	wsMsg := WebSocketMessage{
		Type:      MessageTypeOnlineStatus,
		Data:      statusData,
		Timestamp: time.Now(),
	}

	// Serialize the message once for efficient broadcasting.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal online status: %v", err)
		return
	}

	h.mu.RLock()
	receivers := make([]*Client, 0)
	for clientID, clientSet := range h.Clients {
		if clientID == userID {
			continue
		}
		for client := range clientSet {
			receivers = append(receivers, client)
		}
	}
	h.mu.RUnlock()

	for _, receiver := range receivers {
		select {
		case receiver.Send <- message:
			// Message successfully queued.
		default:
			log.Printf("Skipping online status for user %s (buffer full)", receiver.UserID)
		}
	}
}

// sendOnlineUsersList compiles a list of all currently online users and sends it
// as a single message to a newly connected client (client).
//
// This is essential for initial state synchronization.
func (h *Hub) sendOnlineUsersList(client *Client) {
	h.mu.RLock()
	onlineUsers := make([]OnlineStatusData, 0, len(h.Clients))
	for userID, connections := range h.Clients {
		// Exclude the user receiving the list.
		if userID == client.UserID || len(connections) == 0 {
			continue
		}

		var nickname string
		for c := range connections {
			nickname = c.Nickname
			break
		}

		onlineUsers = append(onlineUsers, OnlineStatusData{
			UserID:   userID,
			Username: nickname,
			IsOnline: true,
		})
	}
	h.mu.RUnlock()

	// 3. Prepare and serialize the list message.
	wsMsg := WebSocketMessage{
		Type:      "online_users_list", // Specific message type for list synchronization
		Data:      onlineUsers,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal online users list: %v", err)
		return
	}

	// 4. Send the message to the target client.
	select {
	case client.Send <- message:
		// Sent successfully.
	default:
		// Handle buffer full scenario for the new connection attempt.
		log.Printf("Failed to send online users list to %s (buffer full/closed)", client.UserID)
	}
}

// IsUserOnline checks the presence of a user in the Hub's Clients map.
//
// This is the thread-safe mechanism for checking if a user has an active
// WebSocket connection managed by this Hub instance.
func (h *Hub) IsUserOnline(userID string) bool {
	// Acquire a Read Lock to safely access the Clients map.
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check for existence; the value is not needed.
	_, exists := h.Clients[userID]
	return exists
}

// GetOnlineUsers returns a list of User IDs for all currently connected clients.
//
// This is used for generating the list of active users for presence synchronization
// or for administrative monitoring.
func (h *Hub) GetOnlineUsers() []string {
	// Acquire a Read Lock for safe map iteration.
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Pre-allocate the slice to the exact size of the map for efficiency.
	users := make([]string, 0, len(h.Clients))

	// Iterate over the map keys (User IDs) and append them to the slice.
	for userID := range h.Clients {
		users = append(users, userID)
	}
	return users
}

// GetOnlineCount returns the total number of active WebSocket connections (unique users).
//
// This is a quick and thread-safe way to get the current size of the connected user base.
func (h *Hub) GetOnlineCount() int {
	// Acquire a Read Lock before accessing the map's length.
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.Clients)
}

// SendChatMessageNotification sends a generic chat message notification (text or image)
// to a specific user.
//
// NOTE: This function uses a generic 'interface{}' for data, which relies on the caller
// to provide the correct underlying structure (e.g., models.MessageWithUser).
func (h *Hub) SendChatMessageNotification(userID string, chatMsgData interface{}) {
	// 1. Prepare the WebSocket message envelope.
	wsMsg := WebSocketMessage{
		Type:      MessageTypeChat, // Generic type for new chat content
		Data:      chatMsgData,
		Timestamp: time.Now(),
	}

	// 2. Serialize the structured message.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message notification: %v", err)
		return
	}

	// 3. Send to the specified user using the robust BroadcastToUser mechanism.
	h.BroadcastToUser(userID, message)
}

// SendChatImageDeleteNotification sends a notification specifically signaling that
// an image file and its associated record have been deleted.
func (h *Hub) SendChatImageDeleteNotification(userID, imageID, messageID string) {
	// 1. Prepare the data payload containing necessary deletion identifiers.
	deleteData := map[string]string{
		"image_id":   imageID,
		"message_id": messageID,
	}

	// 2. Wrap the payload with a dedicated message type.
	wsMsg := WebSocketMessage{
		Type:      "chat_image_deleted", // Specific type for client-side processing
		Data:      deleteData,
		Timestamp: time.Now(),
	}

	// 3. Serialize the notification.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat image delete notification: %v", err)
		return
	}

	// 4. Send the notification to the target user (receiver).
	h.BroadcastToUser(userID, message)
}

// SendNotification sends a real-time notification to a user
func (h *Hub) SendNotification(userID string, notification models.NotificationView) {
	wsMsg := WebSocketMessage{
		Type:      "notification",
		Data:      notification,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	log.Printf("[Hub] Sending notification to user %s: Type=%s, From=%s",
		userID, notification.Type, notification.Nickname)

	h.BroadcastToUser(userID, message)
}
