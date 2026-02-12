package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/utils"
	"social-network/websocket"
)

// ReactionHandler handles like/dislike reactions
type ReactionHandler struct {
	Repo             *repository.ReactionRepository
	PostRepo         *repository.PostRepository
	CommentRepo      *repository.CommentRepository
	NotificationRepo *repository.NotificationRepository
	Hub              *websocket.Hub // ✅ Add Hub reference
}

func NewReactionHandler(
	repo *repository.ReactionRepository,
	postRepo *repository.PostRepository,
	commentRepo *repository.CommentRepository,
	notifRepo *repository.NotificationRepository,
	hub *websocket.Hub, // ✅ Add Hub parameter
) *ReactionHandler {
	return &ReactionHandler{
		Repo:             repo,
		PostRepo:         postRepo,
		CommentRepo:      commentRepo,
		NotificationRepo: notifRepo,
		Hub:              hub,
	}
}

// CreateReact toggles a reaction (like/dislike/love) on a post or comment
// @Summary      React to content
// @Description  Allows an authenticated user to toggle a reaction. If the reaction exists, it is removed; if a different one exists, it is updated. Triggers a real-time notification to the content owner.
// @Tags         Reactions
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Success      200       {array}   models.ReactionWithUser   "Returns the updated list of all reactions for the target"
// @Failure      400       {object}  models.ErrorResponse      "Invalid target_id, target_type, or reaction_type"
// @Failure      401       {object}  models.ErrorResponse      "Unauthorized"
// @Failure      500       {object}  models.ErrorResponse      "Internal Server Error"
// @Router       /api/v1/reactions [post]
func (h *ReactionHandler) CreateReact(w http.ResponseWriter, r *http.Request) {
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
		TargetID     string `json:"target_id"`
		TargetType   string `json:"target_type"`
		ReactionType int    `json:"reaction_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetID == "" || (req.TargetType != "post" && req.TargetType != "comment") || req.ReactionType == 0 {
		utils.ErrorResponse(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Repo.ToggleReaction(user.ID, req.TargetType, req.TargetID, req.ReactionType); err != nil {
		utils.ErrorResponse(w, "Failed to react", http.StatusInternalServerError)
		return
	}

	// ✅ Create notification for the owner of the post/comment
	var notifType string
	if req.ReactionType == 1 {
		notifType = "like"
	} else if req.ReactionType == 2 {
		notifType = "dislike"
	} else if req.ReactionType == 3 {
		notifType = "love"
	}

	if req.TargetType == "post" {
		// Reaction on a post
		if post, err := h.PostRepo.GetByID(req.TargetID); err == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       notifType,
				PostID:     post.ID,
			}
			
			if err := h.NotificationRepo.Create(&n); err != nil {
				log.Printf("[ReactionHandler] Failed to create notification: %v", err)
			} else {
				// ✅ Send real-time notification via WebSocket
				if h.Hub != nil {
					notificationView := models.NotificationView{
						ID:        n.ID,
						Nickname:  user.Nickname,
						Type:      notifType,
						PostID:    post.ID,
						CommentID: nil,
						CreatedAt: n.CreatedAt,
						Read:      false,
						Visible:   true,
					}
					
					log.Printf("[ReactionHandler] Sending %s notification to user %s for post %s", notifType, post.UserID, post.ID)
					h.Hub.SendNotification(post.UserID, notificationView)
				}
			}
		}
	} else {
		// Reaction on a comment
		if comment, err := h.CommentRepo.GetByID(req.TargetID); err == nil && comment != nil && comment.UserID != user.ID {
			n := models.Notification{
				UserID:     comment.UserID,
				FromUserID: user.ID,
				Type:       notifType,
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}
			
			if err := h.NotificationRepo.Create(&n); err != nil {
				log.Printf("[ReactionHandler] Failed to create notification: %v", err)
			} else {
				// ✅ Send real-time notification via WebSocket
				if h.Hub != nil {
					notificationView := models.NotificationView{
						ID:        n.ID,
						Nickname:  user.Nickname,
						Type:      notifType,
						PostID:    comment.PostID,
						CommentID: &comment.ID,
						CreatedAt: n.CreatedAt,
						Read:      false,
						Visible:   true,
					}
					
					log.Printf("[ReactionHandler] Sending %s notification to user %s for comment %s", notifType, comment.UserID, comment.ID)
					h.Hub.SendNotification(comment.UserID, notificationView)
				}
			}
		}
	}

	// Return updated reactions list
	var (
		reactions []models.ReactionWithUser
		err       error
	)
	if req.TargetType == "post" {
		reactions, err = h.Repo.GetReactionsByPostWithUser(req.TargetID)
	} else {
		reactions, err = h.Repo.GetReactionsByCommentWithUser(req.TargetID)
	}
	if err != nil {
		utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, reactions, http.StatusOK)
}
