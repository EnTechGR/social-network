package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// PostHandler handles post related endpoints
type PostHandler struct {
	PostRepo *repository.PostRepository
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(repo *repository.PostRepository) *PostHandler {
	return &PostHandler{PostRepo: repo}
}

// CreatePost creates a new post for the authenticated user
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
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
    CategoryIDs []int  `json:"category_ids"` // Instead of CategoryID
    Title       string `json:"title"`
    Content     string `json:"content"`
}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.CategoryIDs) == 0 || req.Title == "" || req.Content == "" {
    utils.ErrorResponse(w, "At least one category, title and content are required", http.StatusBadRequest)
    return
}

	post := models.Post{
		UserID:     user.ID,
		Title:      &req.Title,
		Content:    &req.Content,
	}

	created, err := h.PostRepo.Create(post, req.CategoryIDs)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, created, http.StatusCreated)
}

// EditPostTitle edits only the title of a post
func (h *PostHandler) EditPostTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}
	var req struct {
		Title *string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Title == nil {
		utils.ErrorResponse(w, "Title is required", http.StatusBadRequest)
		return
	}
	if err := h.PostRepo.UpdatePost(postID, req.Title, nil); err != nil {
		utils.ErrorResponse(w, "Failed to update title", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "title updated"}, http.StatusOK)
}

// EditPostContent edits only the content of a post
func (h *PostHandler) EditPostContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
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
		utils.ErrorResponse(w, "Content is required", http.StatusBadRequest)
		return
	}
	if err := h.PostRepo.UpdatePost(postID, nil, req.Content); err != nil {
		utils.ErrorResponse(w, "Failed to update content", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "content updated"}, http.StatusOK)
}

// DeletePost soft-deletes a post
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}
	if err := h.PostRepo.SoftDeletePost(postID); err != nil {
		utils.ErrorResponse(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
