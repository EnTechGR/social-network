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
	CommentRepo *repository.CommentRepository
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(repo *repository.CommentRepository) *CommentHandler {
	return &CommentHandler{CommentRepo: repo}
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
	if err := h.CommentRepo.SoftDeleteComment(commentID); err != nil {
		utils.ErrorResponse(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
