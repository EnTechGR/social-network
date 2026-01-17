package handlers


import (
   "encoding/json"
   "net/http"


   "social-network/middleware"
   "social-network/models"
   "social-network/repository"
   "social-network/repository/user_repository"
   "social-network/utils"
)


// FollowHandler handles follow-related endpoints.
type FollowHandler struct {
   FollowRepo *repository.FollowRepository
   UserRepo   *user_repository.UserRepository
}


// NewFollowHandler creates a new FollowHandler.
func NewFollowHandler(followRepo *repository.FollowRepository, userRepo *user_repository.UserRepository) *FollowHandler {
   return &FollowHandler{
       FollowRepo: followRepo,
       UserRepo:   userRepo,
   }
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


   if user.IsPrivate {
       utils.ErrorResponse(w, "Private profiles cannot follow in this step", http.StatusForbidden)
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


   utils.JSONResponse(w, relationship, http.StatusCreated)
}


// Unfollow removes an accepted follow relationship.
// @Summary      Unfollow a user
// @Description  Removes an accepted follow relationship for the authenticated user.
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


   if err := h.FollowRepo.DeleteFollow(user.ID, followeeID); err != nil {
       if err == repository.ErrFollowNotFound {
           utils.ErrorResponse(w, "Follow relationship not found", http.StatusNotFound)
       } else {
           utils.ErrorResponse(w, "Failed to unfollow user", http.StatusInternalServerError)
       }
       return
   }


   utils.JSONResponse(w, map[string]string{"status": "unfollowed"}, http.StatusOK)
}


// RemoveFollower removes a follower from the current user's followers list.
// @Summary      Remove a follower
// @Description  Removes an accepted follower relationship for the authenticated user.
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


   if err := h.FollowRepo.RemoveFollower(user.ID, followerID); err != nil {
       if err == repository.ErrFollowNotFound {
           utils.ErrorResponse(w, "Follow relationship not found", http.StatusNotFound)
       } else {
           utils.ErrorResponse(w, "Failed to remove follower", http.StatusInternalServerError)
       }
       return
   }


   utils.JSONResponse(w, map[string]string{"status": "removed"}, http.StatusOK)
}




