package websocket

import (
	"log"
	"net/http"

	"forum/middleware"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from your frontend origin
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8081" || origin == "http://localhost:8080"
	},
}

// HandleWebSocket handles WebSocket connection requests
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		log.Printf("WebSocket connection rejected: user not authenticated")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for user %s: %v", user.ID, err)
		return
	}

	// Create new client
	client := &Client{
		UserID:   user.ID,
		Username: user.Username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	// Register client with hub
	hub.Register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	log.Printf("WebSocket connection established for user %s (%s)", user.Username, user.ID)
}
