package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository/message"
	"social-network/utils"
	"social-network/websocket"
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

// SendMessage sends a new message
// @Summary      Send a private message
// @Description  Persists a message to the database and broadcasts it via WebSocket to the receiver in real-time.
// @Tags         Messaging
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        message  body      models.CreateMessageRequest  true  "Message content and receiver ID"
// @Success      201      {object}  map[string]interface{} "Returns {message: models.Message}"
// @Failure      400      {object}  models.ErrorResponse   "Validation error (e.g., messaging yourself)"
// @Failure      401      {object}  models.ErrorResponse   "Unauthorized"
// @Router       /api/v1/messages [post]
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context, set by authentication middleware.
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body containing message details.
	var req models.CreateMessageRequest
	// Decode the JSON body directly from the request stream.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Clean and normalize input data (security and consistency best practice).
	req.Content = strings.TrimSpace(req.Content)
	req.ReceiverID = strings.TrimSpace(req.ReceiverID)

	// Validate required fields.
	if req.ReceiverID == "" || req.Content == "" {
		utils.ErrorResponse(w, "Receiver ID and content are required", http.StatusBadRequest)
		return
	}

	// Validate content length (data integrity and resource management).
	if utf8.RuneCountInString(req.Content) > 1000 {
		utils.ErrorResponse(w, "Message content cannot exceed 1000 characters", http.StatusBadRequest)
		return
	}

	// Check business logic: Prevent user from messaging themselves.
	if req.ReceiverID == user.ID {
		utils.ErrorResponse(w, "Cannot send message to yourself", http.StatusBadRequest)
		return
	}

	// Enforce social graph rule: users can only chat when at least one accepted follow exists.
	canMessage, err := h.MessageRepo.CanUsersMessage(user.ID, req.ReceiverID)
	if err != nil {
		log.Printf("Failed to verify chat permission: %v", err)
		utils.ErrorResponse(w, "Failed to verify chat permissions", http.StatusInternalServerError)
		return
	}
	if !canMessage {
		utils.ErrorResponse(w, "You can only message users with an accepted follow relationship", http.StatusForbidden)
		return
	}

	// 1. Create message record in the database.
	msg, err := h.MessageRepo.Create(user.ID, req.ReceiverID, req.Content)
	if err != nil {
		log.Printf("Failed to create message: %v", err)
		utils.ErrorResponse(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// 2. Broadcast message via WebSocket when realtime delivery rules are met.
	if h.Hub != nil {
		// Construct the enhanced model required for WebSocket transmission.
		// This includes the sender's nickname for immediate display on the client side.
		msgWithUser := models.MessageWithUser{
			MessageID:      msg.MessageID,
			SenderID:       msg.SenderID,
			SenderNickname: user.Nickname, // Include SenderNickname for the broadcast
			ReceiverID:     msg.ReceiverID,
			Content:        msg.Content,
			CreatedAt:      msg.CreatedAt,
			IsRead:         msg.IsRead,
			// Image field is zero-valued (nil) for standard text messages
		}

		canDeliverRealtime, err := h.MessageRepo.CanDeliverMessageRealtime(user.ID, req.ReceiverID)
		if err != nil {
			log.Printf("Failed to verify realtime chat delivery: %v", err)
			// Best effort fallback: keep sender's active sessions in sync.
			h.Hub.SendChatMessageNotification(user.ID, msgWithUser)
		} else if canDeliverRealtime {
			// Send to both recipient and sender (for multi-device sender sync).
			h.Hub.SendChatMessage(req.ReceiverID, msgWithUser)
		} else {
			// Receiver is not eligible for live delivery; still sync sender sessions.
			h.Hub.SendChatMessageNotification(user.ID, msgWithUser)
		}
	}

	// 3. Send HTTP success response back to the sender.
	// StatusCreated (201) is appropriate for resource creation.
	utils.JSONResponse(w, map[string]interface{}{
		"message": msg,
	}, http.StatusCreated)
}

// GetConversation retrieves messages with a specific user
// @Summary      Get conversation history
// @Description  Retrieves a paginated list of messages between current user and partner. Marks messages as read.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Param        user_id  query     string  true   "ID of the partner"
// @Param        limit    query     int     false  "Max messages (default 10, max 50)"
// @Param        offset   query     int     false  "Pagination offset"
// @Success      200
// @Failure      400      {object}  models.ErrorResponse
// @Router       /api/v1/messages/conversation [get]
func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from query params (required conversation partner ID).
	otherUserID := r.URL.Query().Get("user_id")
	if otherUserID == "" {
		utils.ErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get pagination params from query string.
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default pagination limit
	offset := 0

	// Parse and validate 'limit'. Enforce max limit for resource control.
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Parse and validate 'offset'.
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 1. Get messages for the conversation slice.
	messages, err := h.MessageRepo.GetConversation(user.ID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("Failed to get conversation: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	// 2. Get total message count for accurate pagination metadata.
	total, err := h.MessageRepo.GetConversationCount(user.ID, otherUserID)
	if err != nil {
		log.Printf("Failed to get conversation count: %v", err)
		// If counting fails, default to 0 but allow the primary request to succeed.
		total = 0
	}

	// Determine if there are more messages based on current offset, returned messages, and total count.
	hasMore := (offset + len(messages)) < total

	// 3. Mark messages as read in a non-blocking background goroutine.
	// This improves API response time for the user viewing the conversation.
	go func() {
		err := h.MessageRepo.MarkConversationAsRead(user.ID, otherUserID)
		if err != nil {
			// Log error but do not halt the response flow, as it's an asynchronous side effect.
			log.Printf("Failed to mark conversation as read: %v", err)
		}
	}()

	// 4. Return the paginated messages along with metadata.
	utils.JSONResponse(w, models.MessagesResponse{
		Messages: messages,
		HasMore:  hasMore,
		Total:    total,
	}, http.StatusOK)
}

// GetConversations lists all active chats
// @Summary      List conversations
// @Description  Retrieves all chat partners, last messages, unread counts, and real-time online status.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Router       /api/v1/messages/conversations [get]
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Get conversations metadata from the repository.
	// This usually involves complex SQL queries to identify unique partners
	// and retrieve aggregate data (last message, unread count).
	conversations, err := h.MessageRepo.GetConversations(user.ID)
	if err != nil {
		log.Printf("Failed to get conversations: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve conversations", http.StatusInternalServerError)
		return
	}

	// 2. Augment data with real-time status.
	// Check the WebSocket hub to determine if each conversation partner is currently online.
	if h.Hub != nil {
		for i := range conversations {
			// Mutate the slice element to set the real-time online status flag.
			conversations[i].IsOnline = h.Hub.IsUserOnline(conversations[i].UserID)
		}
	}

	// 3. Return the complete list of conversations to the client.
	utils.JSONResponse(w, models.ConversationsResponse{
		Conversations: conversations,
	}, http.StatusOK)
}

// GetAllUsers retrieves all users with their current online status
// @Summary      List all users
// @Description  Retrieves all registered users (except current) and checks their real-time connection status via the Hub.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/v1/messages/users [get]
func (h *MessageHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Get all users from the repository, excluding the current user.
	users, err := h.MessageRepo.GetAllUsers(user.ID)
	if err != nil {
		log.Printf("Failed to get all users: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// 2. Define a projection struct to augment the data with the IsOnline status.
	// This approach is used because the base User model likely doesn't have IsOnline.
	// Defining it locally ensures the handler explicitly controls the output structure.
	type UserWithOnlineStatus struct {
		ID          string             `json:"id"`
		Nickname    string             `json:"nickname"`
		Email       string             `json:"email"`
		FirstName   string             `json:"first_name"`
		LastName    string             `json:"last_name"`
		DateOfBirth time.Time          `json:"date_of_birth"`
		Gender      string             `json:"gender"`
		Avatar      *models.AvatarInfo `json:"avatar,omitempty"`
		IsOnline    bool               `json:"is_online"` // Added real-time status flag
	}

	// 3. Iterate through retrieved users and add real-time status.
	usersWithStatus := make([]UserWithOnlineStatus, len(users))
	for i, u := range users {
		// Map repository model to the presentation model.
		usersWithStatus[i] = UserWithOnlineStatus{
			ID:          u.ID,
			Nickname:    u.Nickname,
			Email:       u.Email,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DateOfBirth: u.DateOfBirth,
			Gender:      u.Gender,
			Avatar:      u.Avatar,
			IsOnline:    false, // Default to false
		}

		// Check the WebSocket hub for the user's current connection status.
		if h.Hub != nil {
			usersWithStatus[i].IsOnline = h.Hub.IsUserOnline(u.ID)
		}
	}

	// 4. Return the list of users with their online status.
	utils.JSONResponse(w, map[string]interface{}{
		"users": usersWithStatus,
	}, http.StatusOK)
}

// GetUnreadCount gets total unread messages
// @Summary      Unread message count
// @Description  Returns the total count of unread messages across all conversations. Useful for badges.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Router       /api/v1/messages/unread-count [get]
func (h *MessageHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Get the total aggregated unread count from the repository.
	// The repository handles the SQL logic (e.g., SELECT COUNT(*) WHERE receiver_id = ? AND is_read = false).
	count, err := h.MessageRepo.GetTotalUnreadCount(user.ID)
	if err != nil {
		log.Printf("Failed to get unread count: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve unread count", http.StatusInternalServerError)
		return
	}

	// 2. Return the count in a simple JSON format.
	utils.JSONResponse(w, map[string]interface{}{
		"unread_count": count, // Integer representing the total count
	}, http.StatusOK)
}

// MarkAsRead marks a specific message as read
// @Summary      Mark message read
// @Description  Updates a single message status to read. Only the receiver can perform this.
// @Tags         Messaging
// @Security     CookieAuth
// @Param        messageID  path      string  true  "ID of the message"
// @Success      200
// @Failure      403        {object}  models.ErrorResponse "Not the receiver"
// @Failure      404        {object}  models.ErrorResponse "Message not found"
// @Router       /api/v1/messages/read/{messageID} [put]
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Allow both PUT (semantically correct for update) and POST (common for simple actions).
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Extract the message ID from the URL path.
	// Assumes the router pattern ends with the ID (e.g., /mark-as-read/{messageID}).
	pathParts := strings.Split(r.URL.Path, "/")
	// Check for minimum path segments.
	if len(pathParts) < 2 {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID := pathParts[len(pathParts)-1]

	if messageID == "" {
		utils.ErrorResponse(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// 2. Mark message as read in the repository.
	// The repository logic must internally check if the message's ReceiverID matches user.ID.
	err := h.MessageRepo.MarkAsRead(messageID, user.ID)
	if err != nil {
		log.Printf("Failed to mark message as read: %v", err)

		// 3. Granular error handling based on repository-returned error strings.
		// This pattern allows the handler to return specific HTTP status codes (403, 404)
		// without tightly coupling to custom error types, although custom errors are preferred.
		if strings.Contains(err.Error(), "unauthorized") {
			// Message exists but belongs to a different receiver.
			utils.ErrorResponse(w, "Unauthorized", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			// Message ID does not exist.
			utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		} else {
			// General database or other internal error.
			utils.ErrorResponse(w, "Failed to mark message as read", http.StatusInternalServerError)
		}
		return
	}

	// 4. Success response (200 OK for successful update/action).
	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// DeleteMessage deletes a specific message
// @Summary      Delete a message
// @Description  Deletes a message and notifies the receiver via WebSocket. Only the sender can delete.
// @Tags         Messaging
// @Security     CookieAuth
// @Param        messageID  path      string  true  "ID of the message"
// @Success      200
// @Router       /api/v1/messages/{messageID} [delete]
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Extract the message ID from the URL path.
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

	// 2. Get message details before deletion.
	// This step is necessary to retrieve the ReceiverID for the WebSocket notification later.
	msg, err := h.MessageRepo.GetByID(messageID)
	if err != nil {
		log.Printf("Failed to get message: %v", err)
		utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		return
	}

	// 3. Delete message via the repository.
	// The repository function h.MessageRepo.Delete must enforce the authorization rule:
	// only the sender (user.ID) can delete their own message (messageID).
	err = h.MessageRepo.Delete(messageID, user.ID)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		// Granular error handling based on repository output.
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, "Unauthorized: Only sender can delete message", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			// This check is slightly redundant due to the GetByID check above, but provides safety.
			utils.ErrorResponse(w, "Message not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete message", http.StatusInternalServerError)
		}
		return
	}

	// 4. Notify receiver via WebSocket that message was deleted.
	// This ensures the message disappears from the receiver's chat interface in real-time.
	if h.Hub != nil {
		h.Hub.SendMessageDeleteNotification(msg.ReceiverID, messageID, user.ID)
	}

	// 5. Send HTTP success response (200 OK for successful deletion).
	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// GetUsersForChat retrieves potential chat partners
// @Summary      Search chat partners
// @Description  Retrieves users. If all=true, returns everyone. Otherwise, returns only users you haven't chatted with yet.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Param        all   query    bool  false  "Set to true to include users with existing chats"
// @Success      200
// @Failure      401   {object}  models.ErrorResponse
// @Router       /api/v1/messages/search-users [get]
func (h *MessageHandler) GetUsersForChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (ensures authentication).
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Determine retrieval mode based on the 'all' query parameter.
	includeAll := r.URL.Query().Get("all") == "true"

	var users []models.UserWithAvatar
	var err error

	// 1. Conditional data retrieval logic.
	if includeAll {
		// Option A: Get all users (for global user listing).
		users, err = h.MessageRepo.GetAllUsers(user.ID)
	} else {
		// Option B (Default): Get users who are not currently in the user's conversation list (for starting new chats).
		users, err = h.MessageRepo.GetUsersWithoutConversation(user.ID)
	}

	if err != nil {
		log.Printf("Failed to get users for chat: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// UserWithOnlineStatus represents a user with their live presence
	// swagger:model UserWithOnlineStatus
	type UserWithOnlineStatus struct {
		ID          string             `json:"id"`
		Nickname    string             `json:"nickname"`
		Email       string             `json:"email"`
		FirstName   string             `json:"first_name"`
		LastName    string             `json:"last_name"`
		DateOfBirth time.Time          `json:"date_of_birth"`
		Gender      string             `json:"gender"`
		Avatar      *models.AvatarInfo `json:"avatar,omitempty"`
		IsOnline    bool               `json:"is_online"` // Real-time presence status
	}

	usersWithStatus := make([]UserWithOnlineStatus, len(users))
	for i, u := range users {
		// Map user data to the presentation model.
		usersWithStatus[i] = UserWithOnlineStatus{
			ID:          u.ID,
			Nickname:    u.Nickname,
			Email:       u.Email,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DateOfBirth: u.DateOfBirth,
			Gender:      u.Gender,
			Avatar:      u.Avatar,
			IsOnline:    false, // Default status
		}

		// Augment with live status from the WebSocket hub.
		if h.Hub != nil {
			usersWithStatus[i].IsOnline = h.Hub.IsUserOnline(u.ID)
		}
	}

	// 3. Send the final, augmented list of users.
	utils.JSONResponse(w, map[string]interface{}{
		"users": usersWithStatus,
	}, http.StatusOK)
}
