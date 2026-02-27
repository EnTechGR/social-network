package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"social-network/middleware"
	"social-network/repository/group"
	"social-network/utils"
)

// GroupMemberHandler handles group member-related requests
type GroupMemberHandler struct {
	MemberRepo *group.GroupMemberRepository
	GroupRepo  *group.GroupRepository
}

// NewGroupMemberHandler creates a new GroupMemberHandler
func NewGroupMemberHandler(memberRepo *group.GroupMemberRepository, groupRepo *group.GroupRepository) *GroupMemberHandler {
	return &GroupMemberHandler{
		MemberRepo: memberRepo,
		GroupRepo:  groupRepo,
	}
}

// GetGroupMembers retrieves all members of a group
// @Summary      Get group members
// @Description  Retrieves all members of a group (members only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.GroupMemberWithUser
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/members [get]
func (h *GroupMemberHandler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
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
	// URL format: /api/v1/groups/members/{groupID}
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/members/")
	groupID = strings.TrimSpace(groupID)

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if user is a member
	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can view the member list", http.StatusForbidden)
		return
	}

	// Get members
	members, err := h.MemberRepo.GetGroupMembers(groupID)
	if err != nil {
		log.Printf("Failed to get group members: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve members", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"members": members,
	}, http.StatusOK)
}

// RemoveMember removes a member from a group
// @Summary      Remove group member
// @Description  Removes a member from the group (owner only, cannot remove owner)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id        path      string  true  "Group ID"
// @Param        userId    path      string  true  "User ID to remove"
// @Success      200       {object}  map[string]string
// @Failure      400       {object}  models.ErrorResponse
// @Failure      401       {object}  models.ErrorResponse
// @Failure      403       {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/members/{userId} [delete]
func (h *GroupMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract IDs from URL path
	// URL format: /api/v1/groups/members/remove/{groupID}/{userID}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/members/remove/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		utils.ErrorResponse(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	groupID := strings.TrimSpace(parts[0])
	targetUserID := strings.TrimSpace(parts[1])

	if groupID == "" || targetUserID == "" {
		utils.ErrorResponse(w, "Group ID and User ID are required", http.StatusBadRequest)
		return
	}

	// Check if current user is the owner
	isOwner, err := h.GroupRepo.IsUserOwner(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		utils.ErrorResponse(w, "Only the group owner can remove members", http.StatusForbidden)
		return
	}

	// Prevent removing the owner
	isTargetOwner, err := h.GroupRepo.IsUserOwner(groupID, targetUserID)
	if err != nil {
		log.Printf("Failed to check target ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify target user permissions", http.StatusInternalServerError)
		return
	}

	if isTargetOwner {
		utils.ErrorResponse(w, "Cannot remove the group owner", http.StatusBadRequest)
		return
	}

	// Remove member
	if err := h.MemberRepo.RemoveMember(groupID, targetUserID); err != nil {
		log.Printf("Failed to remove member: %v", err)
		utils.ErrorResponse(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Member removed successfully",
	}, http.StatusOK)
}

// LeaveGroup removes the current authenticated user from a group.
// @Summary      Leave group
// @Description  Allows a member to leave a group. Group owner cannot leave; ownership transfer or group deletion is required.
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/groups/leave/{id} [post]
func (h *GroupMemberHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/leave/")
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	isOwner, err := h.GroupRepo.IsUserOwner(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to verify group ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}
	if isOwner {
		utils.ErrorResponse(w, "Group owner cannot leave the group", http.StatusForbidden)
		return
	}

	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to verify membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		utils.ErrorResponse(w, "You are not a member of this group", http.StatusNotFound)
		return
	}

	if err := h.MemberRepo.RemoveMember(groupID, user.ID); err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(w, "You are not a member of this group", http.StatusNotFound)
			return
		}
		log.Printf("Failed to leave group: %v", err)
		utils.ErrorResponse(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Left group successfully",
	}, http.StatusOK)
}

// extractGroupID extracts group ID from URL path
// Example: /api/v1/groups/123/members -> "123"
func extractGroupID(path, prefix, suffix string) string {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	return strings.TrimSpace(path)
}
