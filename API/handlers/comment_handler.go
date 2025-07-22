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
		Content: req.Content,
	}

	created, err := h.CommentRepo.Create(comment)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, created, http.StatusCreated)
}
