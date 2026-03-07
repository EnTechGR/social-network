package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"social-network/pkg/middleware"
	"social-network/pkg/models"
	"social-network/pkg/repository/group"
	"social-network/pkg/utils"
)

// GroupHandler handles group-related requests
type GroupHandler struct {
	GroupRepo  *group.GroupRepository
	MemberRepo *group.GroupMemberRepository
	InviteRepo *group.GroupInviteRepository // ✅ ADD: For sending invites during creation
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(groupRepo *group.GroupRepository, memberRepo *group.GroupMemberRepository, inviteRepo *group.GroupInviteRepository) *GroupHandler {
	return &GroupHandler{
		GroupRepo:  groupRepo,
		MemberRepo: memberRepo,
		InviteRepo: inviteRepo,
	}
}

// CreateGroup creates a new group and optionally invites members
// @Summary      Create a group
// @Description  Creates a new group with the authenticated user as owner and optionally invites members
// @Tags         Groups
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        group  body      object  true  "Group details (title, description, invitees)"
// @Success      201    {object}  models.GroupWithOwner
// @Failure      400    {object}  models.ErrorResponse
// @Failure      401    {object}  models.ErrorResponse
// @Router       /api/v1/groups/create [post]
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Invitees    []string `json:"invitees"` // ✅ NEW: Optional array of user IDs to invite
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize inputs
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		utils.ErrorResponse(w, "Group title is required", http.StatusBadRequest)
		return
	}

	if len(req.Title) > 100 {
		utils.ErrorResponse(w, "Group title cannot exceed 100 characters", http.StatusBadRequest)
		return
	}

	if len(req.Description) > 500 {
		utils.ErrorResponse(w, "Group description cannot exceed 500 characters", http.StatusBadRequest)
		return
	}

	// ✅ NEW: Validate invitees array
	if len(req.Invitees) > 50 {
		utils.ErrorResponse(w, "Cannot invite more than 50 users at once", http.StatusBadRequest)
		return
	}

	// Create group
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	group := models.Group{
		OwnerID:     user.ID,
		Title:       req.Title,
		Description: description,
	}

	createdGroup, err := h.GroupRepo.Create(group)
	if err != nil {
		log.Printf("Failed to create group: %v", err)
		utils.ErrorResponse(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// ✅ NEW: Send invitations to specified users
	invitedCount := 0
	failedInvites := []string{}

	if len(req.Invitees) > 0 && h.InviteRepo != nil {
		for _, inviteeID := range req.Invitees {
			inviteeID = strings.TrimSpace(inviteeID)

			// Skip empty IDs
			if inviteeID == "" {
				continue
			}

			// Skip inviting yourself
			if inviteeID == user.ID {
				continue
			}

			// Check if user is already a member (shouldn't happen, but safety check)
			isMember, err := h.MemberRepo.IsMember(createdGroup.ID, inviteeID)
			if err != nil {
				log.Printf("Failed to check membership for user %s: %v", inviteeID, err)
				failedInvites = append(failedInvites, inviteeID)
				continue
			}

			if isMember {
				continue // Already a member, skip
			}

			// Check for existing pending invite
			hasPending, err := h.InviteRepo.HasPendingInvite(createdGroup.ID, inviteeID)
			if err != nil {
				log.Printf("Failed to check pending invite for user %s: %v", inviteeID, err)
				failedInvites = append(failedInvites, inviteeID)
				continue
			}

			if hasPending {
				continue // Already has pending invite, skip
			}

			// Send invitation
			_, err = h.InviteRepo.Create(createdGroup.ID, user.ID, inviteeID)
			if err != nil {
				log.Printf("Failed to create invite for user %s: %v", inviteeID, err)
				failedInvites = append(failedInvites, inviteeID)
				continue
			}

			invitedCount++
		}
	}

	// Return group with owner details and invitation stats
	groupWithOwner, err := h.GroupRepo.GetByIDWithOwner(createdGroup.ID)
	if err != nil {
		log.Printf("Failed to get group details: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve group details", http.StatusInternalServerError)
		return
	}

	// ✅ NEW: Include invitation statistics in response
	response := map[string]interface{}{
		"group":          groupWithOwner,
		"invites_sent":   invitedCount,
		"invites_failed": len(failedInvites),
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

// BrowseAllGroups retrieves all groups for browsing
// @Summary      Browse all groups
// @Description  Retrieves all groups with owner details and member counts
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Success      200  {array}   models.GroupWithOwner
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/v1/groups [get]
func (h *GroupHandler) BrowseAllGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := h.GroupRepo.GetAll()
	if err != nil {
		log.Printf("Failed to get all groups: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"groups": groups,
	}, http.StatusOK)
}

// GetGroupByID retrieves a specific group
// @Summary      Get group details
// @Description  Retrieves group details with owner information and member count
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  models.GroupWithOwner
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id} [get]
func (h *GroupHandler) GetGroupByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	group, err := h.GroupRepo.GetByIDWithOwner(groupID)
	if err != nil {
		log.Printf("Failed to get group: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve group", http.StatusInternalServerError)
		return
	}

	if group == nil {
		utils.ErrorResponse(w, "Group not found", http.StatusNotFound)
		return
	}

	utils.JSONResponse(w, group, http.StatusOK)
}

// GetMyGroups retrieves all groups the current user is a member of
// @Summary      Get user's groups
// @Description  Retrieves all groups where the authenticated user is a member
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Success      200  {array}   models.GroupWithOwner
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/v1/groups/my-groups [get]
func (h *GroupHandler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := h.GroupRepo.GetUserGroups(user.ID)
	if err != nil {
		log.Printf("Failed to get user groups: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"groups": groups,
	}, http.StatusOK)
}

// UpdateGroup updates a group's title and description
// @Summary      Update group
// @Description  Updates group details (owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Group ID"
// @Param        data  body      object  true  "Updated group details"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id} [put]
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.GroupRepo.IsUserOwner(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can update the group", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		utils.ErrorResponse(w, "Group title is required", http.StatusBadRequest)
		return
	}

	if len(req.Title) > 100 {
		utils.ErrorResponse(w, "Group title cannot exceed 100 characters", http.StatusBadRequest)
		return
	}

	if len(req.Description) > 500 {
		utils.ErrorResponse(w, "Group description cannot exceed 500 characters", http.StatusBadRequest)
		return
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	// Update group
	if err := h.GroupRepo.Update(groupID, req.Title, description); err != nil {
		log.Printf("Failed to update group: %v", err)
		utils.ErrorResponse(w, "Failed to update group", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Group updated successfully",
	}, http.StatusOK)
}

// DeleteGroup deletes a group
// @Summary      Delete group
// @Description  Deletes a group (owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.GroupRepo.IsUserOwner(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can delete the group", http.StatusForbidden)
		return
	}

	// Delete group
	if err := h.GroupRepo.Delete(groupID); err != nil {
		log.Printf("Failed to delete group: %v", err)
		utils.ErrorResponse(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Group deleted successfully",
	}, http.StatusOK)
}
