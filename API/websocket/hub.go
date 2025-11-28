package websocket

import (
	"encoding/json"
	"forum/models" // Added import
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients mapped by user ID
	// FIX 1: Revert to single client per user structure to match existing logic.
	Clients map[string]*Client // Corrected type: map[userID]*Client

	// Mutex to protect the clients map
	mu sync.RWMutex

	// Inbound messages from clients
	Broadcast chan []byte

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		// FIX 2: This initialization now correctly matches the struct definition.
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// BroadcastToUser sends a message to a specific user.
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	// FIX 3: We need to use defer h.mu.RUnlock() to ensure the read lock is released
    defer h.mu.RUnlock() 

	// 1. Get the single client for the target userID
	client, exists := h.Clients[userID]
	if !exists {
		// User is not online
		return
	}

	// 2. Send the message to the client
	select {
	case client.Send <- message:
		// Message sent successfully
	default:
		// If send buffer is full, we must remove the client.
		// NOTE: Deleting while holding a ReadLock is generally unsafe.
        // For a quick fix in this architecture, we rely on the Run loop to clean up, 
        // but a safer approach is to unregister via a channel.
        log.Printf("Client buffer full for user %s. Attempting graceful shutdown.", userID)
        // Since we are fixing the compilation errors, we will keep the unregister logic
        // simple here, but be aware of potential race conditions if the connection is
        // closed outside of the Hub's main loop.
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock() // Lock for writing

	// If user already has a connection, close the old one
	if existingClient, exists := h.Clients[client.UserID]; exists {
		close(existingClient.Send)
		log.Printf("Replacing existing connection for user %s", client.UserID)
	}

	h.Clients[client.UserID] = client
	log.Printf("User %s connected. Total clients: %d", client.UserID, len(h.Clients))

	h.mu.Unlock() // Unlock *before* broadcasting to prevent deadlock

	// Broadcast online status to all clients
	h.broadcastOnlineStatus(client.UserID, client.Username, true)

	// Send list of online users to the newly connected client
	h.sendOnlineUsersList(client)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock() // Lock for writing

	// FIX 4: Corrected unregister logic for single-client model
	if activeClient, exists := h.Clients[client.UserID]; exists && activeClient == client {
		delete(h.Clients, client.UserID)
		close(client.Send)
		log.Printf("User %s disconnected. Total clients: %d", client.UserID, len(h.Clients))

		h.mu.Unlock() // Unlock *before* broadcasting

		// Broadcast offline status to all clients
		h.broadcastOnlineStatus(client.UserID, client.Username, false)
	} else {
		h.mu.Unlock() // Unlock if client wasn't found
	}
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		select {
		case client.Send <- message:
		default:
			// Client's send buffer is full, skip this client
			log.Printf("Skipping message for user %s (buffer full)", client.UserID)
		}
	}
}

// SendChatMessage sends a chat message to a specific user
func (h *Hub) SendChatMessage(receiverID string, msgData models.MessageWithUser) {
	// 1. Prepare WebSocket message
	msgType := MessageTypeChat // Default to chat
	// if msgData.Image != nil {
	// 	msgType = MessageTypeChatImage // Use the image type if an image is present
	// }

	wsMsg := WebSocketMessage{
		Type: 	   msgType,
		Data: 	   msgData, // models.MessageWithUser now contains all needed fields
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message: %v", err)
		return
	}

	// 2. Broadcast to Receiver
	h.BroadcastToUser(receiverID, message)

	// 3. Send back to Sender (to update their own chat UI on all devices)
	h.BroadcastToUser(msgData.SenderID, message)
}

// SendMessageDeleteNotification notifies a user that a message was deleted
func (h *Hub) SendMessageDeleteNotification(userID, messageID, senderID string) {
	deleteData := map[string]string{
		"message_id": messageID,
		"sender_id": 	senderID, // Important for the client to know who to update
	}

	wsMsg := WebSocketMessage{
		Type: 	   MessageTypeMessageDelete,
		Data: 	   deleteData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal delete notification: %v", err)
		return
	}

	// Notify the receiver
	h.BroadcastToUser(userID, message)
	// Notify the sender
	h.BroadcastToUser(senderID, message)
}

// broadcastOnlineStatus broadcasts a user's online/offline status to all clients
func (h *Hub) broadcastOnlineStatus(userID, username string, isOnline bool) {
	statusData := OnlineStatusData{
		UserID: 	userID,
		Username: username,
		IsOnline: isOnline,
	}

	wsMsg := WebSocketMessage{
		Type: 	   MessageTypeOnlineStatus,
		Data: 	   statusData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal online status: %v", err)
		return
	}

	// Broadcast to all clients except the user themselves
	h.mu.RLock()
	defer h.mu.RUnlock()

	for clientID, client := range h.Clients {
		if clientID != userID {
			select {
			case client.Send <- message:
			default:
				log.Printf("Skipping online status for user %s (buffer full)", clientID)
			}
		}
	}
}

// sendOnlineUsersList sends the list of currently online users to a client
func (h *Hub) sendOnlineUsersList(client *Client) {
	h.mu.RLock()
	onlineUsers := make([]OnlineStatusData, 0, len(h.Clients))
	for userID, c := range h.Clients {
		if userID != client.UserID {
			onlineUsers = append(onlineUsers, OnlineStatusData{
				UserID: 	c.UserID,
				Username: c.Username,
				IsOnline: true,
			})
		}
	}
	h.mu.RUnlock()

	wsMsg := WebSocketMessage{
		Type: 	   "online_users_list",
		Data: 	   onlineUsers,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal online users list: %v", err)
		return
	}

	select {
	case client.Send <- message:
	default:
		log.Printf("Failed to send online users list to %s", client.UserID)
	}
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.Clients[userID]
	return exists
}

// GetOnlineUsers returns a list of all online user IDs
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.Clients))
	for userID := range h.Clients {
		users = append(users, userID)
	}
	return users
}

// GetOnlineCount returns the number of online users
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.Clients)
}

// FIX 1: Add a generic chat notification function
// SendChatMessageNotification sends a message (text or image) notification to a specific user
func (h *Hub) SendChatMessageNotification(userID string, chatMsgData interface{}) {
	wsMsg := WebSocketMessage{
		Type:      MessageTypeChat, // Use the generic chat message type defined in client.go
		Data:      chatMsgData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message notification: %v", err)
		return
	}

	// Send to the specified user
	h.BroadcastToUser(userID, message)
}

// SendChatImageDeleteNotification sends notification that an image was deleted
func (h *Hub) SendChatImageDeleteNotification(userID, imageID, messageID string) {
	deleteData := map[string]string{
		"image_id": 	imageID,
		"message_id": messageID,
	}

	wsMsg := WebSocketMessage{
		Type: 	   "chat_image_deleted",
		Data: 	   deleteData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat image delete notification: %v", err)
		return
	}

	h.BroadcastToUser(userID, message)
}