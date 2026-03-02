package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"social-network/middleware"
	"social-network/repository"
	"social-network/repository/group"
	"social-network/utils"
	"social-network/websocket"
)

// GroupInviteHandler handles group invitation requests
type GroupInviteHandler struct {
	InviteRepo       *group.GroupInviteRepository
	MemberRepo       *group.GroupMemberRepository
	GroupRepo        *group.GroupRepository
	NotificationRepo *repository.NotificationRepository
	Hub              *websocket.Hub
}

// NewGroupInviteHandler creates a new GroupInviteHandler
func NewGroupInviteHandler(
	inviteRepo *group.GroupInviteRepository,
	memberRepo *group.GroupMemberRepository,
	groupRepo *group.GroupRepository,
	notificationRepo *repository.NotificationRepository,
	hub *websocket.Hub,
) *GroupInviteHandler {
	return &GroupInviteHandler{
		InviteRepo:       inviteRepo,
		MemberRepo:       memberRepo,
		GroupRepo:        groupRepo,
		NotificationRepo: notificationRepo,
		Hub:              hub,
	}
}

// InviteUser invites a user to join a group
// @Summary      Invite user to group
// @Description  Sends an invitation to a user to join the group (members only)
// @Tags         Groups
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Group ID"
// @Param        data  body      object  true  "User ID to invite"
// @Success      201   {object}  models.GroupInvite
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/invite [post]
func (h *GroupInviteHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	// URL format: /api/v1/groups/invite/{groupID}
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/invite/")
	groupID = strings.TrimSpace(groupID)

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)

	if req.UserID == "" {
		utils.ErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if current user is a member (members can invite)
	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can invite others", http.StatusForbidden)
		return
	}

	// Check if target user is already a member
	isAlreadyMember, err := h.MemberRepo.IsMember(groupID, req.UserID)
	if err != nil {
		log.Printf("Failed to check target membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify target user status", http.StatusInternalServerError)
		return
	}

	if isAlreadyMember {
		utils.ErrorResponse(w, "User is already a member of this group", http.StatusBadRequest)
		return
	}

	// Check for existing pending invite
	hasPending, err := h.InviteRepo.HasPendingInvite(groupID, req.UserID)
	if err != nil {
		log.Printf("Failed to check pending invite: %v", err)
		utils.ErrorResponse(w, "Failed to verify invite status", http.StatusInternalServerError)
		return
	}

	if hasPending {
		utils.ErrorResponse(w, "User already has a pending invite to this group", http.StatusBadRequest)
		return
	}

	// Create invitation
	invite, err := h.InviteRepo.Create(groupID, user.ID, req.UserID)
	if err != nil {
		log.Printf("Failed to create invite: %v", err)
		utils.ErrorResponse(w, "Failed to send invitation", http.StatusInternalServerError)
		return
	}

	createAndPushNotification(h.NotificationRepo, h.Hub, req.UserID, user.ID, user.Nickname, "group_invite")

	utils.JSONResponse(w, invite, http.StatusCreated)
}

// GetMyInvites retrieves all pending invites for the current user
// @Summary      Get my group invites
// @Description  Retrieves all pending group invitations for the authenticated user
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Success      200  {array}   models.GroupInviteWithDetails
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/v1/groups/invites [get]
func (h *GroupInviteHandler) GetMyInvites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invites, err := h.InviteRepo.GetPendingInvitesForUser(user.ID)
	if err != nil {
		log.Printf("Failed to get invites: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve invitations", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"invites": invites,
	}, http.StatusOK)
}

// AcceptInvite accepts a group invitation
// @Summary      Accept group invite
// @Description  Accepts a group invitation and adds user to the group
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Invite ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/groups/invites/{id}/accept [put]
func (h *GroupInviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract invite ID from URL
	// Path: /api/v1/groups/invites/accept/{inviteID}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.ErrorResponse(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	inviteID := parts[5] // api/v1/groups/invites/accept/{inviteID}

	if inviteID == "" {
		utils.ErrorResponse(w, "Invite ID is required", http.StatusBadRequest)
		return
	}

	if err := h.InviteRepo.AcceptInviteForUser(inviteID, user.ID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			utils.ErrorResponse(w, "Invitation not found", http.StatusNotFound)
			return
		case errors.Is(err, group.ErrInviteNotRecipient):
			utils.ErrorResponse(w, "You can only accept your own invitations", http.StatusForbidden)
			return
		case errors.Is(err, group.ErrInviteAlreadyHandled):
			utils.ErrorResponse(w, "This invitation has already been responded to", http.StatusBadRequest)
			return
		default:
			log.Printf("Failed to accept invite: %v", err)
			utils.ErrorResponse(w, "Failed to accept invitation", http.StatusInternalServerError)
			return
		}
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Invitation accepted successfully",
	}, http.StatusOK)
}

// DeclineInvite declines a group invitation
// @Summary      Decline group invite
// @Description  Declines a group invitation
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Invite ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/groups/invites/{id}/decline [put]
func (h *GroupInviteHandler) DeclineInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract invite ID from URL
	// Path: /api/v1/groups/invites/decline/{inviteID}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.ErrorResponse(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	inviteID := parts[5] // api/v1/groups/invites/decline/{inviteID}

	if inviteID == "" {
		utils.ErrorResponse(w, "Invite ID is required", http.StatusBadRequest)
		return
	}

	// Get invite details
	invite, err := h.InviteRepo.GetByID(inviteID)
	if err != nil {
		log.Printf("Failed to get invite: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve invitation", http.StatusInternalServerError)
		return
	}

	if invite == nil {
		utils.ErrorResponse(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Verify the invite is for the current user
	if invite.ToUserID != user.ID {
		utils.ErrorResponse(w, "You can only decline your own invitations", http.StatusForbidden)
		return
	}

	// Check if invite is still pending
	if invite.Status != "pending" {
		utils.ErrorResponse(w, "This invitation has already been responded to", http.StatusBadRequest)
		return
	}

	// Update invite status
	if err := h.InviteRepo.UpdateStatus(inviteID, "declined"); err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(w, "Invitation not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to update invite status: %v", err)
		utils.ErrorResponse(w, "Failed to decline invitation", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Invitation declined",
	}, http.StatusOK)
}
