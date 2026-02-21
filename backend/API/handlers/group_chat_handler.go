package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository/message"
	"social-network/utils"
	"social-network/websocket"
)

// GroupChatHandler handles group chat HTTP endpoints.
type GroupChatHandler struct {
	MessageRepo *message.MessageRepository
	Hub         *websocket.Hub
}

// NewGroupChatHandler creates a new GroupChatHandler.
func NewGroupChatHandler(messageRepo *message.MessageRepository, hub *websocket.Hub) *GroupChatHandler {
	return &GroupChatHandler{
		MessageRepo: messageRepo,
		Hub:         hub,
	}
}

// SendGroupMessage creates and broadcasts a new group chat message.
func (h *GroupChatHandler) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/v1/groups/chat/send/"))
	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	isMember, err := h.MessageRepo.IsGroupMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to verify group membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify group membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		utils.ErrorResponse(w, "Only group members can chat in this room", http.StatusForbidden)
		return
	}

	var req models.CreateGroupMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		utils.ErrorResponse(w, "Message content is required", http.StatusBadRequest)
		return
	}

	msg, err := h.MessageRepo.CreateGroupMessage(groupID, user.ID, req.Content)
	if err != nil {
		log.Printf("Failed to create group message: %v", err)
		if strings.Contains(err.Error(), "cannot exceed 1000") {
			utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		utils.ErrorResponse(w, "Failed to send group message", http.StatusInternalServerError)
		return
	}

	if h.Hub != nil {
		memberIDs, err := h.MessageRepo.GetGroupMemberIDs(groupID)
		if err != nil {
			log.Printf("Failed to fetch group members for websocket fanout: %v", err)
		} else {
			h.Hub.SendGroupChatMessage(memberIDs, *msg)
		}
	}

	utils.JSONResponse(w, map[string]interface{}{
		"message": msg,
	}, http.StatusCreated)
}

// GetGroupMessages returns paginated group chat history for a member.
func (h *GroupChatHandler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/v1/groups/chat/messages/"))
	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	isMember, err := h.MessageRepo.IsGroupMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to verify group membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify group membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		utils.ErrorResponse(w, "Only group members can view this chat", http.StatusForbidden)
		return
	}

	limit := 50
	offset := 0
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsedLimit, err := strconv.Atoi(rawLimit); err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}
	if rawOffset := strings.TrimSpace(r.URL.Query().Get("offset")); rawOffset != "" {
		if parsedOffset, err := strconv.Atoi(rawOffset); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := h.MessageRepo.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		log.Printf("Failed to get group messages: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve group messages", http.StatusInternalServerError)
		return
	}

	total, err := h.MessageRepo.GetGroupMessageCount(groupID)
	if err != nil {
		log.Printf("Failed to count group messages: %v", err)
		total = 0
	}

	utils.JSONResponse(w, models.GroupMessagesResponse{
		Messages: messages,
		HasMore:  offset+len(messages) < total,
		Total:    total,
	}, http.StatusOK)
}
