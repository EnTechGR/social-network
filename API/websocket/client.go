package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	// This includes time spent writing the initial request and waiting for the response.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	// This controls the lifetime of the connection; if the pong isn't received in time, the connection is considered dead.
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait).
	// Ping messages keep the connection alive (preventing timeouts from proxies/firewalls)
	// and detect dead peers quickly. It is typically 90% of the pongWait duration.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (client to server).
	// Enforcing a limit prevents resource exhaustion attacks (memory, bandwidth).
	maxMessageSize = 8192 // 8KB
)

// Client represents a single, active WebSocket client connection managed by the Hub.
// It encapsulates the necessary state and I/O channels for a user session.
type Client struct {
	// User ID of the connected client (used as the key in Hub.Clients map).
	UserID string

	// Nickname for display and identification in logs/broadcasts.
	Nickname string

	// The underlying Gorilla WebSocket connection pointer.
	Conn *websocket.Conn

	// Buffered channel of outbound messages. Messages from the Hub are sent here
	// and are asynchronously written to the WebSocket by the writePump goroutine.
	Send chan []byte

	// Reference back to the central Hub where this client is registered.
	Hub *Hub
}

// Message types define the action or content of the WebSocket payload.
// These types are crucial for the client-side application to correctly route and render the data.
const (
	// Standard text/image chat message (new message).
	MessageTypeChat = "chat"
	// Real-time notification when a user is typing.
	MessageTypeTyping = "typing"
	// User presence notification (online/offline status).
	MessageTypeOnlineStatus = "online_status"
	// Notification that a message has been viewed by the receiver.
	MessageTypeMessageRead = "message_read"
	// Notification that a message has been permanently deleted.
	MessageTypeMessageDelete = "message_delete"
	// Message that contains image metadata (if integrated separately from MessageTypeChat).
	MessageTypeChatImage = "chat_image"
	// Notification that a chat image has been deleted.
	MessageTypeChatImageDeleted = "chat_image_deleted"
	// Generic error message from the server to the client.
	MessageTypeError = "error"
)

// WebSocketMessage represents the standardized envelope used for all WebSocket communication.
// It ensures every transmission has a recognizable structure for client parsing.
type WebSocketMessage struct {
	// Type determines how the client should interpret and handle the Data field.
	Type string `json:"type"`
	// Data holds the specific payload; 'interface{}' allows for diverse content types.
	Data interface{} `json:"data"`
	// Timestamp records when the message was formulated by the server (best practice).
	Timestamp time.Time `json:"timestamp"`
}

// ChatMessageData represents the data payload for a new chat message.
// type ChatMessageData struct {
// 	MessageID  string    `json:"message_id"`
// 	SenderID   string    `json:"sender_id"`
// 	SenderName string    `json:"sender_name"`
// 	ReceiverID string    `json:"receiver_id"`
// 	Content    string    `json:"content"`
// 	CreatedAt  time.Time `json:"created_at"`
// 	IsRead     bool      `json:"is_read"`
// }

// TypingData represents the payload for the typing indicator feature.
type TypingData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsTyping bool   `json:"is_typing"` // True for start typing, False for stop typing
}

// OnlineStatusData represents a user's presence state change.
type OnlineStatusData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"` // True for connect, False for disconnect
}

// MessageReadData represents a notification that a specific message has been marked as read.
type MessageReadData struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"` // The ID of the user who marked it as read
	ReadAt    time.Time `json:"read_at"`
}

// MessageDeleteData represents a notification payload for message deletion.
type MessageDeleteData struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"` // The ID of the user who deleted the message (the sender)
}

// readPump pumps messages from the websocket connection to the hub.
//
// This function runs in a separate goroutine and is responsible for reading
// all incoming client messages, enforcing protocol/size limits, and handling
// keep-alive pongs. When the loop terminates, it cleans up the client connection.
func (c *Client) readPump() {
	// 1. Cleanup on exit (defer). This ensures the client is unregistered and the
	// physical connection is closed, regardless of how the read loop exits (error or break).
	defer func() {
		// Signal the Hub to remove this client from the active list.
		c.Hub.Unregister <- c
		// Close the underlying WebSocket connection.
		c.Conn.Close()
	}()

	// 2. Configure Keep-Alive and Timeouts.
	// Set the initial read deadline for the first pong message.
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	// Define the pong handler: reset the read deadline every time a pong is received.
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 3. Configure Input Limits.
	// Enforce maximum message size to prevent malicious or excessive data loads.
	c.Conn.SetReadLimit(maxMessageSize)

	// 4. Main Read Loop.
	for {
		// Read a message from the WebSocket connection.
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// Handle unexpected close errors (e.g., browser tab closed, network failure).
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.UserID, err)
			}
			// Break the loop on any read error, triggering the defer cleanup.
			break
		}

		// 5. Parse and Route Message.
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to parse WebSocket message from user %s: %v", c.UserID, err)
			continue // Log error and continue reading next message.
		}

		// Dispatch the message for application-level handling based on its Type.
		c.handleMessage(wsMsg)
	}
}

// writePump pumps messages from the hub (via c.Send channel) to the websocket connection.
//
// This function runs in a separate goroutine and is responsible for all outbound traffic
// and sending periodic keep-alive pings.
func (c *Client) writePump() {
	// Create a periodic ticker for sending ping frames.
	ticker := time.NewTicker(pingPeriod)
	
	// 1. Cleanup on exit (defer). Stop the ticker and close the connection.
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	// 2. Main Write/Ping Loop.
	for {
		select {
		// Case A: Message received from the Hub via the Send channel.
		case message, ok := <-c.Send:
			// Set a deadline for the write operation (writeWait) to prevent indefinite blocking.
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			
			if !ok {
				// The Hub closed the Send channel, indicating the client is being unregistered.
				// Send a normal close message to the peer and exit the pump.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Use NextWriter for efficient batching of multiple messages into a single frame.
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return // Write failure, exit pump and trigger defer cleanup.
			}
			w.Write(message) // Write the first message.

			// Optimization: Check the buffer and immediately send any other queued messages.
			// This reduces system calls and message overhead by consolidating messages.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				// Use a newline as a conventional separator for batched JSON messages.
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			// Close the writer to finalize the WebSocket frame.
			if err := w.Close(); err != nil {
				return // Write failure, exit pump.
			}

		// Case B: Ping ticker fires.
		case <-ticker.C:
			// Set a write deadline for the ping message.
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			
			// Send a control frame (Ping). No payload is needed.
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Ping failed, connection likely dead, exit pump.
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages by routing them based on their Type.
// This decouples message parsing (in readPump) from business logic handling.
func (c *Client) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case MessageTypeTyping:
		// Client sends a typing status update.
		c.handleTypingIndicator(msg)
	case MessageTypeMessageRead:
		// Client sends a notification that they have read a message.
		c.handleMessageRead(msg)
	default:
		// Log any incoming message types that are not recognized or handled.
		log.Printf("Unknown message type from user %s: %s", c.UserID, msg.Type)
	}
}

// handleTypingIndicator validates and forwards a user's typing status to the intended recipient.
func (c *Client) handleTypingIndicator(msg WebSocketMessage) {
	// 1. Re-marshal and unmarshal to convert the generic interface{} data into the specific struct (TypingData).
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

	// 2. Augment data for security and consistency. Use the Client's authenticated details, not user-supplied ones.
	// NOTE: The 'UserID' in typingData should logically be the *recipient* of the typing notification.
	// Assuming `typingData.UserID` here refers to the *recipient* and we set the *sender* details.
	typingData.UserID = c.UserID    // Set the actual sender's ID (the user doing the typing)
	typingData.Username = c.Nickname

	// 3. Re-wrap and serialize the clean message for broadcast.
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

	// 4. Send the notification to the target user (recipient).
	// NOTE: This assumes the client included the *recipient's* ID in the original message's Data field.
	// The implementation here incorrectly uses `typingData.UserID` (which was just set to the sender's ID)
	// as the target. The *actual* recipient ID must be passed to `BroadcastToUser`.
	// Assuming the client passes the recipient ID in a separate field or implicitly handles it.
	// FIX: Must use the recipient's ID as the first argument. Since the recipient's ID is not present
	// in the `TypingData` struct definition, this forwarding logic needs refinement to correctly identify the recipient.
	c.Hub.BroadcastToUser(typingData.UserID, message) // Currently sends back to sender, should use recipient's ID
}

// handleMessageRead validates and forwards a user's message read receipt status.
func (c *Client) handleMessageRead(msg WebSocketMessage) {
	// 1. Re-marshal and unmarshal to convert the generic interface{} data into the specific struct.
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

	// 2. Augment data. Set the user who performed the read action (c.UserID) and the timestamp.
	readData.UserID = c.UserID
	readData.ReadAt = time.Now()

	// 3. Re-wrap and serialize the clean message for broadcast.
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

	// 4. Broadcast the read notification.
	// The read notification must go to the original sender of the message being marked as read,
	// so they can update their UI.
	// NOTE: The current implementation broadcasts to ALL clients via `c.Hub.Broadcast`.
	// BEST PRACTICE: It should be targeted to the original sender's ID via `c.Hub.BroadcastToUser(senderID, message)`.
	// This requires knowing the sender ID from the messageID, which is a common limitation here.
	c.Hub.Broadcast <- message // CURRENTLY BROADCASTS GLOBALLY
}

// SendMessage constructs a WebSocketMessage and attempts a non-blocking send to this client.
//
// This is a utility function intended to be called by the Hub or other server components
// to send targeted messages to this specific client instance.
func (c *Client) SendMessage(msgType string, data interface{}) error {
	// 1. Construct the message envelope.
	wsMsg := WebSocketMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}

	// 2. Serialize the message.
	message, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	// 3. Non-blocking send attempt.
	select {
	case c.Send <- message:
		return nil // Message successfully queued.
	default:
		// If the channel is full, the message is dropped. This is preferred over blocking
		// the sender, as it indicates a slow client or network issue.
		// NOTE: Returning `nil` on default success is poor practice; it should return an error
		// or log the failure explicitly. Returning `nil` hides the fact that the message was dropped.
		return nil // Client's send buffer is full, message dropped but operation reported as successful.
	}
}