package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"social-network/middleware"
	"social-network/repository"
	"social-network/repository/group"
	"social-network/utils"
	"social-network/websocket"
)

// GroupJoinRequestHandler handles group join request operations
type GroupJoinRequestHandler struct {
	RequestRepo      *group.GroupJoinRequestRepository
	MemberRepo       *group.GroupMemberRepository
	GroupRepo        *group.GroupRepository
	NotificationRepo *repository.NotificationRepository
	Hub              *websocket.Hub
}

// NewGroupJoinRequestHandler creates a new GroupJoinRequestHandler
func NewGroupJoinRequestHandler(
	requestRepo *group.GroupJoinRequestRepository,
	memberRepo *group.GroupMemberRepository,
	groupRepo *group.GroupRepository,
	notificationRepo *repository.NotificationRepository,
	hub *websocket.Hub,
) *GroupJoinRequestHandler {
	return &GroupJoinRequestHandler{
		RequestRepo:      requestRepo,
		MemberRepo:       memberRepo,
		GroupRepo:        groupRepo,
		NotificationRepo: notificationRepo,
		Hub:              hub,
	}
}

// RequestToJoin creates a join request for a group
// @Summary      Request to join group
// @Description  Creates a join request to join a group
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      201  {object}  models.GroupJoinRequest
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/request [post]
func (h *GroupJoinRequestHandler) RequestToJoin(w http.ResponseWriter, r *http.Request) {
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
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/request/")
	groupID = strings.TrimSpace(groupID)

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Verify group exists
	group, err := h.GroupRepo.GetByID(groupID)
	if err != nil {
		log.Printf("Failed to get group: %v", err)
		utils.ErrorResponse(w, "Failed to verify group", http.StatusInternalServerError)
		return
	}

	if group == nil {
		utils.ErrorResponse(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if user is already a member
	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if isMember {
		utils.ErrorResponse(w, "You are already a member of this group", http.StatusBadRequest)
		return
	}

	// Check for existing pending request
	hasPending, err := h.RequestRepo.HasPendingRequest(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check pending request: %v", err)
		utils.ErrorResponse(w, "Failed to verify request status", http.StatusInternalServerError)
		return
	}

	if hasPending {
		utils.ErrorResponse(w, "You already have a pending request for this group", http.StatusBadRequest)
		return
	}

	// Create join request
	request, err := h.RequestRepo.Create(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to create join request: %v", err)
		utils.ErrorResponse(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	createAndPushNotification(h.NotificationRepo, h.Hub, group.OwnerID, user.ID, user.Nickname, "group_join_request")

	utils.JSONResponse(w, request, http.StatusCreated)
}

// GetPendingRequests retrieves all pending join requests for a group (owner only)
// @Summary      Get pending join requests
// @Description  Retrieves all pending join requests for a group (owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.GroupJoinRequestWithDetails
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/requests [get]
func (h *GroupJoinRequestHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/requests/")
	groupID = strings.TrimSpace(groupID)

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if user is the owner
	isOwner, err := h.GroupRepo.IsUserOwner(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can view join requests", http.StatusForbidden)
		return
	}

	// Get pending requests
	requests, err := h.RequestRepo.GetPendingRequestsForGroup(groupID)
	if err != nil {
		log.Printf("Failed to get join requests: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve requests", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"requests": requests,
	}, http.StatusOK)
}

// ApproveRequest approves a join request (owner only)
// @Summary      Approve join request
// @Description  Approves a join request and adds the user to the group (owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Request ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/groups/requests/{id}/approve [put]
func (h *GroupJoinRequestHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract request ID from URL
	// Path: /api/v1/groups/requests/approve/{requestID}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.ErrorResponse(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	requestID := parts[5] // api/v1/groups/requests/approve/{requestID}

	if requestID == "" {
		utils.ErrorResponse(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	// Get request details
	request, err := h.RequestRepo.GetByID(requestID)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve request", http.StatusInternalServerError)
		return
	}

	if request == nil {
		utils.ErrorResponse(w, "Request not found", http.StatusNotFound)
		return
	}

	// Check if current user is the owner
	isOwner, err := h.GroupRepo.IsUserOwner(request.GroupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can approve requests", http.StatusForbidden)
		return
	}

	// Check if request is still pending
	if request.Status != "pending" {
		utils.ErrorResponse(w, "This request has already been processed", http.StatusBadRequest)
		return
	}

	// Update request status
	if err := h.RequestRepo.UpdateStatus(requestID, "approved"); err != nil {
		log.Printf("Failed to update request status: %v", err)
		utils.ErrorResponse(w, "Failed to approve request", http.StatusInternalServerError)
		return
	}

	// Add user to group
	isMember, err := h.MemberRepo.IsMember(request.GroupID, request.UserID)
	if err != nil {
		log.Printf("Failed to verify member before add: %v", err)
		utils.ErrorResponse(w, "Failed to add member to group", http.StatusInternalServerError)
		return
	}
	if !isMember {
		if err := h.MemberRepo.AddMember(request.GroupID, request.UserID); err != nil {
			log.Printf("Failed to add member: %v", err)
			utils.ErrorResponse(w, "Failed to add member to group", http.StatusInternalServerError)
			return
		}
	}

	createAndPushNotification(h.NotificationRepo, h.Hub, request.UserID, user.ID, user.Nickname, "group_join_decision")

	utils.JSONResponse(w, map[string]string{
		"message": "Request approved successfully",
	}, http.StatusOK)
}

// DenyRequest denies a join request (owner only)
// @Summary      Deny join request
// @Description  Denies a join request (owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Request ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/groups/requests/{id}/deny [put]
func (h *GroupJoinRequestHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract request ID from URL
	// Path: /api/v1/groups/requests/deny/{requestID}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.ErrorResponse(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	requestID := parts[5] // api/v1/groups/requests/deny/{requestID}

	if requestID == "" {
		utils.ErrorResponse(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	// Get request details
	request, err := h.RequestRepo.GetByID(requestID)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve request", http.StatusInternalServerError)
		return
	}

	if request == nil {
		utils.ErrorResponse(w, "Request not found", http.StatusNotFound)
		return
	}

	// Check if current user is the owner
	isOwner, err := h.GroupRepo.IsUserOwner(request.GroupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can deny requests", http.StatusForbidden)
		return
	}

	// Check if request is still pending
	if request.Status != "pending" {
		utils.ErrorResponse(w, "This request has already been processed", http.StatusBadRequest)
		return
	}

	// Update request status
	if err := h.RequestRepo.UpdateStatus(requestID, "denied"); err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(w, "Request not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to update request status: %v", err)
		utils.ErrorResponse(w, "Failed to deny request", http.StatusInternalServerError)
		return
	}

	createAndPushNotification(h.NotificationRepo, h.Hub, request.UserID, user.ID, user.Nickname, "group_join_decision")

	utils.JSONResponse(w, map[string]string{
		"message": "Request denied",
	}, http.StatusOK)
}
