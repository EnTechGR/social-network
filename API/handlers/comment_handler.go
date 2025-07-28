package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// CommentHandler handles comment related endpoints
type CommentHandler struct {
	CommentRepo      *repository.CommentRepository
	PostRepo         *repository.PostRepository
	NotificationRepo *repository.NotificationRepository
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(
	commentRepo *repository.CommentRepository,
	postRepo *repository.PostRepository,
	notifRepo *repository.NotificationRepository,
) *CommentHandler {
	return &CommentHandler{
		CommentRepo:      commentRepo,
		PostRepo:         postRepo,
		NotificationRepo: notifRepo,
	}
}

// CreateComment creates a new comment on a post for the authenticated user
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PostID == "" || req.Content == "" {
		utils.ErrorResponse(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostID:  req.PostID,
		UserID:  user.ID,
		Content: &req.Content,
	}

	created, err := h.CommentRepo.Create(comment)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// notify post owner about new comment
	if post, err := h.PostRepo.GetByID(req.PostID); err == nil && post != nil && post.UserID != user.ID {
		n := models.Notification{
			UserID:     post.UserID,
			FromUserID: user.ID,
			Type:       "comment",
			PostID:     req.PostID,
			CommentID:  &created.ID,
		}
		_ = h.NotificationRepo.Create(n)
	}

	utils.JSONResponse(w, created, http.StatusCreated)
}

// EditComment edits a comment's content
func (h *CommentHandler) EditComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	commentID := utils.GetLastPathParam(r)
	if commentID == "" {
		utils.ErrorResponse(w, "Missing comment ID", http.StatusBadRequest)
		return
	}

	// Check ownership - only comment owner can edit
	comment, err := h.CommentRepo.GetCommentByID(commentID)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			utils.ErrorResponse(w, "Comment not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to retrieve comment", http.StatusInternalServerError)
		}
		return
	}
	if comment.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only edit your own comments", http.StatusForbidden)
		return
	}

	var req struct {
		Content *string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Content == nil {
		utils.ErrorResponse(w, "Nothing to update", http.StatusBadRequest)
		return
	}
	if err := h.CommentRepo.UpdateComment(commentID, req.Content); err != nil {
		utils.ErrorResponse(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	if comment, err := h.CommentRepo.GetByID(commentID); err == nil && comment != nil {
		if post, err2 := h.PostRepo.GetByID(comment.PostID); err2 == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       "edit_comment",
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}
			_ = h.NotificationRepo.Create(n)
		}
	}
	utils.JSONResponse(w, map[string]string{"status": "updated"}, http.StatusOK)
}

// DeleteComment soft-deletes a comment
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	commentID := utils.GetLastPathParam(r)
	if commentID == "" {
		utils.ErrorResponse(w, "Missing comment ID", http.StatusBadRequest)
		return
	}

	// Check ownership - only comment owner can delete
	comment, err := h.CommentRepo.GetCommentByID(commentID)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			utils.ErrorResponse(w, "Comment not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to retrieve comment", http.StatusInternalServerError)
		}
		return
	}
	if comment.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only delete your own comments", http.StatusForbidden)
		return
	}

	if comment != nil {
		if post, err := h.PostRepo.GetByID(comment.PostID); err == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       "delete_comment",
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}
			_ = h.NotificationRepo.Create(n)
		}
	}
	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
