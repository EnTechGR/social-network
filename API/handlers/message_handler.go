package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"forum/middleware"
	"forum/models"
	"forum/repository/message"
	"forum/utils"
	"forum/websocket"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	MessageRepo *message.MessageRepository
	Hub         *websocket.Hub // ✅ ADD WebSocket Hub
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(messageRepo *message.MessageRepository, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{
		MessageRepo: messageRepo,
		Hub:         hub,
	}
}

// SendMessage handles sending a new message
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim content
	req.Content = strings.TrimSpace(req.Content)
	req.ReceiverID = strings.TrimSpace(req.ReceiverID)

	// Validate request
	if req.ReceiverID == "" || req.Content == "" {
		utils.ErrorResponse(w, "Receiver ID and content are required", http.StatusBadRequest)
		return
	}

	// Validate content length
	if len(req.Content) > 1000 {
		utils.ErrorResponse(w, "Message content cannot exceed 1000 characters", http.StatusBadRequest)
		return
	}

	// Check if trying to message self
	if req.ReceiverID == user.ID {
		utils.ErrorResponse(w, "Cannot send message to yourself", http.StatusBadRequest)
		return
	}

	// Create message
	msg, err := h.MessageRepo.Create(user.ID, req.ReceiverID, req.Content)
	if err != nil {
		log.Printf("Failed to create message: %v", err)
		utils.ErrorResponse(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// ✅ Broadcast message via WebSocket to the receiver
	if h.Hub != nil {
		chatData := websocket.ChatMessageData{
			MessageID:  msg.MessageID,
			SenderID:   msg.SenderID,
			SenderName: user.Username,
			ReceiverID: msg.ReceiverID,
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt,
			IsRead:     msg.IsRead,
		}
		h.Hub.SendChatMessage(req.ReceiverID, chatData)
	}

	utils.JSONResponse(w, map[string]interface{}{
		"message": msg,
		"success": true,
	}, http.StatusCreated)
}

// GetConversation retrieves messages between current user and another user
func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from query params
	otherUserID := r.URL.Query().Get("user_id")
	if otherUserID == "" {
		utils.ErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get pagination params
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get messages
	messages, err := h.MessageRepo.GetConversation(user.ID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("Failed to get conversation: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	// Get total count
	total, err := h.MessageRepo.GetConversationCount(user.ID, otherUserID)
	if err != nil {
		log.Printf("Failed to get conversation count: %v", err)
		total = 0
	}

	// Determine if there are more messages
	hasMore := (offset + len(messages)) < total

	// Mark messages as read (messages sent to current user)
	go func() {
		err := h.MessageRepo.MarkConversationAsRead(user.ID, otherUserID)
		if err != nil {
			log.Printf("Failed to mark conversation as read: %v", err)
		}
	}()

	utils.JSONResponse(w, models.MessagesResponse{
		Messages: messages,
		HasMore:  hasMore,
		Total:    total,
	}, http.StatusOK)
}

// GetConversations retrieves all conversations for the current user
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get conversations
	conversations, err := h.MessageRepo.GetConversations(user.ID)
	if err != nil {
		log.Printf("Failed to get conversations: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve conversations", http.StatusInternalServerError)
		return
	}

	// ✅ Set online status from WebSocket manager
	if h.Hub != nil {
		for i := range conversations {
			conversations[i].IsOnline = h.Hub.IsUserOnline(conversations[i].UserID)
		}
	}

	utils.JSONResponse(w, models.ConversationsResponse{
		Conversations: conversations,
	}, http.StatusOK)
}

// GetAllUsers retrieves all users for the chat list
func (h *MessageHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all users except current user
	users, err := h.MessageRepo.GetAllUsers(user.ID)
	if err != nil {
		log.Printf("Failed to get all users: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"users": users,
	}, http.StatusOK)
}

// GetUnreadCount retrieves total unread message count
func (h *MessageHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get total unread count
	count, err := h.MessageRepo.GetTotalUnreadCount(user.ID)
	if err != nil {
		log.Printf("Failed to get unread count: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve unread count", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"unread_count": count,
	}, http.StatusOK)
}

// MarkAsRead marks a specific message as read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get message ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID := pathParts[len(pathParts)-1]

	if messageID == "" {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// Mark message as read
	err := h.MessageRepo.MarkAsRead(messageID, user.ID)
	if err != nil {
		log.Printf("Failed to mark message as read: %v", err)
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, "Unauthorized", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to mark message as read", http.StatusInternalServerError)
		}
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// DeleteMessage deletes a message (sender only)
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get message ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID := pathParts[len(pathParts)-1]

	if messageID == "" {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// Get message to find receiver ID before deletion
	msg, err := h.MessageRepo.GetByID(messageID)
	if err != nil {
		log.Printf("Failed to get message: %v", err)
		utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		return
	}

	// Delete message
	err = h.MessageRepo.Delete(messageID, user.ID)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, "Unauthorized: Only sender can delete message", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete message", http.StatusInternalServerError)
		}
		return
	}

	// ✅ Notify receiver via WebSocket that message was deleted
	if h.Hub != nil {
		h.Hub.SendMessageDeleteNotification(msg.ReceiverID, messageID, user.ID)
	}

	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// GetUsersForChat retrieves users to start new conversations with
func (h *MessageHandler) GetUsersForChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check query param to determine which users to get
	includeAll := r.URL.Query().Get("all") == "true"

	var users []models.User
	var err error

	if includeAll {
		// Get all users
		users, err = h.MessageRepo.GetAllUsers(user.ID)
	} else {
		// Get only users without existing conversations
		users, err = h.MessageRepo.GetUsersWithoutConversation(user.ID)
	}

	if err != nil {
		log.Printf("Failed to get users for chat: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// ✅ Add online status from WebSocket manager
	type UserWithOnlineStatus struct {
		models.User
		IsOnline bool `json:"is_online"`
	}

	usersWithStatus := make([]UserWithOnlineStatus, len(users))
	for i, u := range users {
		usersWithStatus[i] = UserWithOnlineStatus{
			User:     u,
			IsOnline: false,
		}
		if h.Hub != nil {
			usersWithStatus[i].IsOnline = h.Hub.IsUserOnline(u.ID)
		}
	}

	utils.JSONResponse(w, map[string]interface{}{
		"users": usersWithStatus,
	}, http.StatusOK)
}