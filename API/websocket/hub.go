package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients mapped by user ID
	Clients map[string]*Client

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
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
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

	h.mu.Unlock() // ✅ Unlock *before* broadcasting to prevent deadlock

	// Broadcast online status to all clients
	h.broadcastOnlineStatus(client.UserID, client.Username, true)

	// Send list of online users to the newly connected client
	h.sendOnlineUsersList(client)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock() // Lock for writing

	if _, exists := h.Clients[client.UserID]; exists {
		delete(h.Clients, client.UserID)
		close(client.Send)
		log.Printf("User %s disconnected. Total clients: %d", client.UserID, len(h.Clients))

		h.mu.Unlock() // ✅ Unlock *before* broadcasting

		// Broadcast offline status to all clients
		h.broadcastOnlineStatus(client.UserID, client.Username, false)
	} else {
		h.mu.Unlock() // ✅ Unlock if client wasn't found
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

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, exists := h.Clients[userID]; exists {
		select {
		case client.Send <- message:
		default:
			log.Printf("Skipping message for user %s (buffer full)", userID)
		}
	}
}

// SendChatMessage sends a chat message to a specific user
func (h *Hub) SendChatMessage(receiverID string, msgData ChatMessageData) {
	wsMsg := WebSocketMessage{
		Type:      MessageTypeChat,
		Data:      msgData,
		Timestamp: msgData.CreatedAt,
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message: %v", err)
		return
	}

	h.BroadcastToUser(receiverID, message)
}

// SendMessageDeleteNotification notifies a user that a message was deleted
func (h *Hub) SendMessageDeleteNotification(receiverID, messageID, senderID string) {
	deleteData := MessageDeleteData{
		MessageID: messageID,
		UserID:    senderID,
	}

	wsMsg := WebSocketMessage{
		Type:      MessageTypeMessageDelete,
		Data:      deleteData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal delete notification: %v", err)
		return
	}

	h.BroadcastToUser(receiverID, message)
}

// broadcastOnlineStatus broadcasts a user's online/offline status to all clients
func (h *Hub) broadcastOnlineStatus(userID, username string, isOnline bool) {
	statusData := OnlineStatusData{
		UserID:   userID,
		Username: username,
		IsOnline: isOnline,
	}

	wsMsg := WebSocketMessage{
		Type:      MessageTypeOnlineStatus,
		Data:      statusData,
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
				UserID:   c.UserID,
				Username: c.Username,
				IsOnline: true,
			})
		}
	}
	h.mu.RUnlock()

	wsMsg := WebSocketMessage{
		Type:      "online_users_list",
		Data:      onlineUsers,
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