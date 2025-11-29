package websocket

import (
	"encoding/json"
	"forum/models" // Added import
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and handles message broadcasting.
// It acts as the central router for all WebSocket traffic.
type Hub struct {
	// Registered clients mapped by user ID.
	// This map stores the active WebSocket connection for each authenticated user.
	Clients map[string]*Client // Key: UserID (string), Value: The active Client connection struct

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
		Clients: make(map[string]*Client),
		// Use a moderately sized buffer for the broadcast channel (256) to handle bursts of messages.
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// BroadcastToUser sends a raw message payload to a single, specific user's active connection.
// This function is thread-safe as it uses a Read Lock (RLock) to access the Clients map.
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	// Acquire a Read Lock to safely read the 'Clients' map.
	h.mu.RLock()
	// Ensure the Read Lock is released when the function exits, regardless of the path.
	defer h.mu.RUnlock() 

	// 1. Attempt to retrieve the client connection for the target user ID.
	client, exists := h.Clients[userID]
	if !exists {
		// User is not currently connected via WebSocket; silently drop the message.
		return
	}

	// 2. Non-blocking send attempt to the client's outgoing 'Send' channel.
	select {
	case client.Send <- message:
		// Message sent successfully to the client's outgoing buffer.
	default:
		// The client's 'Send' channel buffer is full. This indicates the client is
		// reading too slowly or is unresponsive, suggesting a dead connection.
		log.Printf("Client buffer full for user %s. Attempting graceful shutdown.", userID)
		
		// NOTE ON RACE CONDITION: Sending to the Unregister channel from within a
		// Broadcast function (which holds an RLock) is safe, but the actual removal
		// of the client from the map must happen in the Hub's main Run loop, which
		// holds the Write Lock (Lock). This pattern avoids deadlocks.
		
		// Unregistering here is a standard way to handle a slow client,
		// relying on the Hub's main loop to clean up the client connection.
		h.Unregister <- client 
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
	// 1. Acquire Write Lock. All modifications to the Clients map must be protected.
	h.mu.Lock()

	// 2. Handle concurrent connections: If the user already has a connection, close the old one.
	if existingClient, exists := h.Clients[client.UserID]; exists {
		// Closing the Send channel signals the old client's read/write loops to terminate.
		close(existingClient.Send)
		log.Printf("Replacing existing connection for user %s", client.UserID)
	}

	// 3. Register the new client connection.
	h.Clients[client.UserID] = client
	log.Printf("User %s connected. Total clients: %d", client.UserID, len(h.Clients))

	// 4. Release Write Lock. Crucial to allow other goroutines (like BroadcastToUser) to proceed.
	h.mu.Unlock() 

	// 5. Broadcast status updates (performed outside the lock).
	// Notify all other connected clients that this user is now online.
	h.broadcastOnlineStatus(client.UserID, client.Username, true)

	// Send list of online users to the newly connected client for initial state synchronization.
	h.sendOnlineUsersList(client)
}

// unregisterClient removes a client from the hub's Clients map and cleans up its resources.
//
// This operation is critical for maintaining accurate user presence and preventing resource leaks.
func (h *Hub) unregisterClient(client *Client) {
	// 1. Acquire Write Lock. Modification of the Clients map must be protected.
	h.mu.Lock()

	// 2. Safely remove the client.
	// Only proceed if the client exists AND the client being unregistered is the currently active connection
	// for that UserID. This prevents removing a newer connection if the unregister request is from an old one.
	if activeClient, exists := h.Clients[client.UserID]; exists && activeClient == client {
		// Remove the client from the map.
		delete(h.Clients, client.UserID)
		
		// Close the client's outgoing Send channel. This signals the client's goroutines
		// (writer loop) to terminate cleanly, releasing the associated WebSocket connection.
		close(client.Send)
		
		log.Printf("User %s disconnected. Total clients: %d", client.UserID, len(h.Clients))

		// 3. Release Write Lock. Crucial to allow other goroutines to proceed.
		h.mu.Unlock()

		// 4. Broadcast status update (performed outside the lock).
		// Notify all remaining connected clients that this user is now offline.
		h.broadcastOnlineStatus(client.UserID, client.Username, false)
	} else {
		// The client was already replaced by a newer connection or never fully registered.
		// Simply unlock and do nothing.
		h.mu.Unlock()
	}
}

// broadcastMessage routes a raw message payload to the outgoing channel of EVERY connected client.
//
// This is typically used for general, non-targeted messages (e.g., global announcements)
// or for real-time presence updates (handled by auxiliary functions).
func (h *Hub) broadcastMessage(message []byte) {
	// Acquire a Read Lock to safely iterate over the Clients map.
	h.mu.RLock()
	// Ensure the Read Lock is released when the function completes.
	defer h.mu.RUnlock()

	// Iterate through all currently connected clients.
	for _, client := range h.Clients {
		// Use a non-blocking select statement to send the message.
		select {
		case client.Send <- message:
			// Message successfully buffered for the client's writer goroutine.
		default:
			// The client's Send channel is full. Instead of blocking the Hub's main
			// broadcast loop (which would stop all traffic), we skip this client.
			// The client's own read loop should eventually detect the problem and
			// trigger an unregister request if the connection is dead.
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

	// 3. Broadcast the status update (thread-safe iteration).
	h.mu.RLock()
	defer h.mu.RUnlock() // Release Read Lock when iteration is complete

	// Iterate through all connected clients.
	for clientID, client := range h.Clients {
		// IMPORTANT: Do not send the status update to the user whose status just changed.
		if clientID != userID {
			select {
			case client.Send <- message:
				// Message successfully queued.
			default:
				// Skip client if buffer is full to prevent blocking the status broadcast loop.
				log.Printf("Skipping online status for user %s (buffer full)", clientID)
			}
		}
	}
}

// sendOnlineUsersList compiles a list of all currently online users and sends it
// as a single message to a newly connected client (client).
//
// This is essential for initial state synchronization.
func (h *Hub) sendOnlineUsersList(client *Client) {
	// 1. Acquire Read Lock to safely access the Clients map for iteration.
	h.mu.RLock()
	
	// Pre-allocate slice capacity for efficiency.
	onlineUsers := make([]OnlineStatusData, 0, len(h.Clients))
	
	// Build the list of currently active users.
	for userID, c := range h.Clients {
		// Exclude the client receiving the list.
		if userID != client.UserID {
			onlineUsers = append(onlineUsers, OnlineStatusData{
				UserID:   c.UserID,
				Username: c.Username,
				IsOnline: true, // They are currently online
			})
		}
	}
	
	// 2. Release the lock immediately after reading is complete.
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
        userID, notification.Type, notification.Username)

    h.BroadcastToUser(userID, message)
}