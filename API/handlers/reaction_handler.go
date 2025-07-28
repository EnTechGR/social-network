package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// ReactionHandler handles like/dislike reactions
type ReactionHandler struct {
	Repo             *repository.ReactionRepository
	PostRepo         *repository.PostRepository
	CommentRepo      *repository.CommentRepository
	NotificationRepo *repository.NotificationRepository
}

func NewReactionHandler(repo *repository.ReactionRepository, postRepo *repository.PostRepository, commentRepo *repository.CommentRepository, notifRepo *repository.NotificationRepository) *ReactionHandler {
	return &ReactionHandler{Repo: repo, PostRepo: postRepo, CommentRepo: commentRepo, NotificationRepo: notifRepo}
}

// React toggles a reaction on a post or comment for the authenticated user
func (h *ReactionHandler) CreateReact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	// Create notification for the owner of the post/comment
	var notifType string
	if req.ReactionType == 1 {
		notifType = "like"
	} else {
		notifType = "dislike"
	}
	if req.TargetType == "post" {
		if post, err := h.PostRepo.GetByID(req.TargetID); err == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       notifType,
				PostID:     post.ID,
			}
			_ = h.NotificationRepo.Create(n)
		}
	} else {
		if comment, err := h.CommentRepo.GetByID(req.TargetID); err == nil && comment != nil && comment.UserID != user.ID {
			n := models.Notification{
				UserID:     comment.UserID,
				FromUserID: user.ID,
				Type:       notifType,
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}
			_ = h.NotificationRepo.Create(n)
		}
	}

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
