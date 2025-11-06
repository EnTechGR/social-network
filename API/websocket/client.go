package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192 // 8KB
)

// Client represents a WebSocket client connection
type Client struct {
	// User ID of the connected client
	UserID string

	// Username for display
	Username string

	// The websocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// Reference to the hub
	Hub *Hub
}

// Message types
const (
	MessageTypeChat          = "chat"
	MessageTypeTyping        = "typing"
	MessageTypeOnlineStatus  = "online_status"
	MessageTypeMessageRead   = "message_read"
	MessageTypeMessageDelete = "message_delete"
	MessageTypeError         = "error"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ChatMessageData represents chat message data
type ChatMessageData struct {
	MessageID  string    `json:"message_id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsRead     bool      `json:"is_read"`
}

// TypingData represents typing indicator data
type TypingData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsTyping bool   `json:"is_typing"`
}

// OnlineStatusData represents online status data
type OnlineStatusData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"`
}

// MessageReadData represents message read notification
type MessageReadData struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

// MessageDeleteData represents message deletion notification
type MessageDeleteData struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.UserID, err)
			}
			break
		}

		// Parse the incoming message
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to parse WebSocket message from user %s: %v", c.UserID, err)
			continue
		}

		// Handle different message types
		c.handleMessage(wsMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case MessageTypeTyping:
		// Handle typing indicator
		c.handleTypingIndicator(msg)
	case MessageTypeMessageRead:
		// Handle message read notification
		c.handleMessageRead(msg)
	default:
		log.Printf("Unknown message type from user %s: %s", c.UserID, msg.Type)
	}
}

// handleTypingIndicator broadcasts typing status to the recipient
func (c *Client) handleTypingIndicator(msg WebSocketMessage) {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		log.Printf("Failed to marshal typing data: %v", err)
		return
	}

	var typingData TypingData
	if err := json.Unmarshal(dataBytes, &typingData); err != nil {
		log.Printf("Failed to unmarshal typing data: %v", err)
		return
	}

	// Broadcast typing status to the recipient
	typingData.UserID = c.UserID
	typingData.Username = c.Username

	wsMsg := WebSocketMessage{
		Type:      MessageTypeTyping,
		Data:      typingData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal typing message: %v", err)
		return
	}

	// Send to specific user (implement in hub)
	c.Hub.BroadcastToUser(typingData.UserID, message)
}

// handleMessageRead handles message read notifications
func (c *Client) handleMessageRead(msg WebSocketMessage) {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		log.Printf("Failed to marshal message read data: %v", err)
		return
	}

	var readData MessageReadData
	if err := json.Unmarshal(dataBytes, &readData); err != nil {
		log.Printf("Failed to unmarshal message read data: %v", err)
		return
	}

	readData.UserID = c.UserID
	readData.ReadAt = time.Now()

	wsMsg := WebSocketMessage{
		Type:      MessageTypeMessageRead,
		Data:      readData,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal message read notification: %v", err)
		return
	}

	// Broadcast read notification (implement in hub)
	c.Hub.Broadcast <- message
}

// SendMessage sends a message to this client
func (c *Client) SendMessage(msgType string, data interface{}) error {
	wsMsg := WebSocketMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}

	message, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- message:
		return nil
	default:
		return nil // Client's send buffer is full, skip this message
	}
}
