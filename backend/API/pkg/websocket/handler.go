package websocket

import (
	"log"
	"net/http"

	"social-network/pkg/middleware"
	"social-network/pkg/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// ReadBufferSize and WriteBufferSize define the initial size of the I/O buffers.
	// 1024 bytes is a common, moderate default size.
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin is a security measure to prevent Cross-Site WebSocket Hijacking (CSWSH).
	CheckOrigin: func(r *http.Request) bool {
		// Only allow WebSocket connections originating from trusted frontend domains (CORS policy).
		origin := r.Header.Get("Origin")
		// NOTE: In a production environment, this list should use fully qualified, non-localhost domains.
		return origin == "http://localhost:8081" || origin == "http://localhost:8080"
	},
}

// HandleWebSocket handles incoming HTTP requests destined to become WebSocket connections.
//
// This function performs initial authentication, upgrades the connection protocol,
// creates a new Client instance, and initiates the read/write message pumps.
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 1. Authentication Check: Get the authenticated user from the context, which must be
	// set by preceding authentication middleware (e.g., cookie or session verification).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		log.Printf("WebSocket connection rejected: user not authenticated")
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Protocol Upgrade: Upgrade the standard HTTP connection to the WebSocket protocol.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for user %s: %v", user.ID, err)
		return // Cannot proceed; silently return.
	}

	// 3. Client Creation: Construct a new Client object associated with the user and the connection.
	client := &Client{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Conn:     conn,
		// Send channel is buffered to handle messages arriving faster than the socket can write them.
		Send: make(chan []byte, 256),
		Hub:  hub,
	}

	// 4. Client Registration: Queue the client for registration in the Hub's central loop.
	// This unbuffered channel send will block until the Hub's Run loop receives it.
	hub.Register <- client

	// 5. Start Goroutines: Initiate the concurrent message handling loops.
	// These two goroutines (readPump and writePump) manage the connection's I/O lifecycle.
	go client.writePump() // Handles messages from the Hub to the WebSocket.
	go client.readPump()  // Handles incoming messages from the WebSocket to the Hub/Broadcast channel.

	log.Printf("WebSocket connection established for user %s (%s)", user.Nickname, user.ID)
}
