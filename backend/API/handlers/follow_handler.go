package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/repository/user_repository"
	"social-network/utils"
	"social-network/websocket"
)

// FollowHandler handles follow-related endpoints.
type FollowHandler struct {
	FollowRepo       *repository.FollowRepository
	UserRepo         *user_repository.UserRepository
	NotificationRepo *repository.NotificationRepository
	Hub              *websocket.Hub
}

// NewFollowHandler creates a new FollowHandler.
func NewFollowHandler(
	followRepo *repository.FollowRepository,
	userRepo *user_repository.UserRepository,
	notificationRepo *repository.NotificationRepository,
	hub *websocket.Hub,
) *FollowHandler {
	return &FollowHandler{
		FollowRepo:       followRepo,
		UserRepo:         userRepo,
		NotificationRepo: notificationRepo,
		Hub:              hub,
	}
}

func (h *FollowHandler) createAndPushFollowNotification(targetUserID, fromUserID, fromNickname, notificationType string) {
	if h.NotificationRepo == nil {
		return
	}

	n := models.Notification{
		UserID:     targetUserID,
		FromUserID: fromUserID,
		Type:       notificationType,
	}

	if err := h.NotificationRepo.Create(&n); err != nil {
		log.Printf("[FollowHandler] Failed to create %s notification for user %s: %v", notificationType, targetUserID, err)
		return
	}

	if h.Hub == nil {
		return
	}

	h.Hub.SendNotification(targetUserID, models.NotificationView{
		ID:        n.ID,
		Nickname:  fromNickname,
		Type:      notificationType,
		PostID:    "",
		CommentID: nil,
		CreatedAt: n.CreatedAt,
		Read:      false,
		Visible:   true,
	})
}

// FollowPublicUser creates an accepted follow relationship when both users are public.
// @Summary      Follow a public user
// @Description  Creates an accepted follow relationship when both the follower and followee are public profiles.
// @Tags         Follow
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        data body object true "Follow target {\"followee_id\": \"uuid\"}"
// @Success      201  {object}  models.FollowRelationship
// @Failure      400  {object}  models.ErrorResponse "Invalid request or self-follow"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      403  {object}  models.ErrorResponse "Only public profiles are supported"
// @Failure      404  {object}  models.ErrorResponse "User not found"
// @Failure      409  {object}  models.ErrorResponse "Already following"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/follow [post]
func (h *FollowHandler) FollowPublicUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FolloweeID string `json:"followee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.FolloweeID == "" {
		utils.ErrorResponse(w, "Followee ID is required", http.StatusBadRequest)
		return
	}

	followee, err := h.UserRepo.GetByID(req.FolloweeID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			utils.ErrorResponse(w, "User not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to load user", http.StatusInternalServerError)
		}
		return
	}

	var relationship *models.FollowRelationship
	if followee.IsPrivate {
		relationship, err = h.FollowRepo.CreateFollowRequest(user.ID, followee.ID)
	} else {
		relationship, err = h.FollowRepo.CreateAcceptedFollow(user.ID, followee.ID)
	}
	if err != nil {
		switch err {
		case repository.ErrFollowSelf:
			utils.ErrorResponse(w, "You cannot follow yourself", http.StatusBadRequest)
		case repository.ErrFollowAlreadyExists:
			utils.ErrorResponse(w, "Already following", http.StatusConflict)
		case repository.ErrFollowBlocked:
			utils.ErrorResponse(w, "Follow not allowed", http.StatusForbidden)
		default:
			utils.ErrorResponse(w, "Failed to follow user", http.StatusInternalServerError)
		}
		return
	}

	if relationship != nil {
		switch relationship.Status {
		case "pending":
			h.createAndPushFollowNotification(followee.ID, user.ID, user.Nickname, "follow_request")
		case "accepted":
			h.createAndPushFollowNotification(followee.ID, user.ID, user.Nickname, "follow_accept")
		}
	}

	utils.JSONResponse(w, relationship, http.StatusCreated)
}

// AcceptFollowRequest accepts a pending follow request for the current user.
// @Summary      Accept a follow request
// @Description  Accepts a pending follow request for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Follower ID"
// @Success      200  {object}  map[string]string "status: accepted"
// @Failure      400  {object}  models.ErrorResponse "Missing follower ID"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse "Follow request not found"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/follow/accept/{id} [put]
func (h *FollowHandler) AcceptFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followerID := utils.GetLastPathParam(r)
	if followerID == "" {
		utils.ErrorResponse(w, "Missing follower ID", http.StatusBadRequest)
		return
	}

	if err := h.FollowRepo.AcceptFollowRequest(followerID, user.ID); err != nil {
		if err == repository.ErrFollowNotFound {
			utils.ErrorResponse(w, "Follow request not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to accept follow request", http.StatusInternalServerError)
		}
		return
	}

	h.createAndPushFollowNotification(followerID, user.ID, user.Nickname, "follow_accept")

	utils.JSONResponse(w, map[string]string{"status": "accepted"}, http.StatusOK)
}

// GetFollowRequests returns pending and accepted follow requests for the current user.
// @Summary      Get follow requests
// @Description  Returns pending and accepted follow requests for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  map[string][]models.FollowRelationship
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/follow/requests [get]
func (h *FollowHandler) GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pending, err := h.FollowRepo.GetByFolloweeAndStatus(user.ID, "pending")
	if err != nil {
		utils.ErrorResponse(w, "Failed to load follow requests", http.StatusInternalServerError)
		return
	}

	accepted, err := h.FollowRepo.GetByFolloweeAndStatus(user.ID, "accepted")
	if err != nil {
		utils.ErrorResponse(w, "Failed to load follow requests", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Pending  []models.FollowRelationship `json:"pending"`
		Accepted []models.FollowRelationship `json:"accepted"`
	}{
		Pending:  pending,
		Accepted: accepted,
	}

	utils.JSONResponse(w, resp, http.StatusOK)
}

// Unfollow removes an accepted or pending follow relationship.
// @Summary      Unfollow a user
// @Description  Removes an accepted or pending follow relationship for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Followee ID"
// @Success      200  {object}  map[string]string "status: unfollowed"
// @Failure      400  {object}  models.ErrorResponse "Missing followee ID"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse "Follow relationship not found"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/follow/delete/{id} [delete]
func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followeeID := utils.GetLastPathParam(r)
	if followeeID == "" {
		utils.ErrorResponse(w, "Missing followee ID", http.StatusBadRequest)
		return
	}

	if err := h.FollowRepo.DeleteRelationship(user.ID, followeeID); err != nil {
		if err == repository.ErrFollowNotFound {
			utils.ErrorResponse(w, "Follow relationship not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to unfollow user", http.StatusInternalServerError)
		}
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "unfollowed"}, http.StatusOK)
}

// GetFollowers returns accepted followers for the current user.
// @Summary      Get followers
// @Description  Returns accepted followers for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  map[string][]models.FollowRelationship
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/followers [get]
func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followers, err := h.FollowRepo.GetByFolloweeAndStatus(user.ID, "accepted")
	if err != nil {
		utils.ErrorResponse(w, "Failed to load followers", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Followers []models.FollowRelationship `json:"followers"`
	}{
		Followers: followers,
	}

	utils.JSONResponse(w, resp, http.StatusOK)
}

// GetPendingFollowRequests returns pending follow requests for the current user.
// @Summary      Get pending follow requests
// @Description  Returns pending follow requests for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  map[string][]models.FollowRelationship
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/follow/requests/pending [get]
func (h *FollowHandler) GetPendingFollowRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pending, err := h.FollowRepo.GetByFolloweeAndStatus(user.ID, "pending")
	if err != nil {
		utils.ErrorResponse(w, "Failed to load follow requests", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Pending []models.FollowRelationship `json:"pending"`
	}{
		Pending: pending,
	}

	utils.JSONResponse(w, resp, http.StatusOK)
}

// RemoveFollower removes a follower from the current user's followers list.
// @Summary      Remove a follower
// @Description  Removes an accepted or pending follower relationship for the authenticated user.
// @Tags         Follow
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Follower ID"
// @Success      200  {object}  map[string]string "status: removed"
// @Failure      400  {object}  models.ErrorResponse "Missing follower ID"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse "Follow relationship not found"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/followers/delete/{id} [delete]
func (h *FollowHandler) RemoveFollower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followerID := utils.GetLastPathParam(r)
	if followerID == "" {
		utils.ErrorResponse(w, "Missing follower ID", http.StatusBadRequest)
		return
	}

	if err := h.FollowRepo.DeleteRelationship(followerID, user.ID); err != nil {
		if err == repository.ErrFollowNotFound {
			utils.ErrorResponse(w, "Follow relationship not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to remove follower", http.StatusInternalServerError)
		}
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "removed"}, http.StatusOK)
}
