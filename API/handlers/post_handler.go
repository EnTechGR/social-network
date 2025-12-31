package handlers

import (
	"encoding/json"
	"net/http"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/utils"
)

// PostHandler handles post related endpoints
type PostHandler struct {
	PostRepo *repository.PostRepository
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(repo *repository.PostRepository) *PostHandler {
	return &PostHandler{PostRepo: repo}
}

// CreatePost creates a new social-network post
// @Summary      Create a post
// @Description  Creates a new post for the authenticated user with a title, content, and multiple categories.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        post  body      object  true  "Post Data (category_ids, title, content)"
// @Success      201   {object}  models.Post
// @Failure      400   {object}  models.ErrorResponse "Invalid request body or missing fields"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/posts [post]
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
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

// EditPostTitle updates the title of a specific post
// @Summary      Edit post title
// @Description  Allows the post owner to update only the title of their post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        title body      object  true  "New Title"
// @Success      200
// @Failure      403   {object}  models.ErrorResponse "Forbidden: Not the owner"
// @Failure      404   {object}  models.ErrorResponse "Post not found"
// @Router       /api/posts/{id}/title [put]
func (h *PostHandler) EditPostTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Check ownership - only post owner can edit
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		if err == repository.ErrPostNotFound {
			utils.ErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}
	if post.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only edit your own posts", http.StatusForbidden)
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

// EditPostContent updates the text content of a specific post
// @Summary      Edit post content
// @Description  Allows the post owner to update only the main text content of their post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Post ID"
// @Param        content body      object  true  "New Content"
// @Success      200
// @Failure      403     {object}  models.ErrorResponse "Forbidden: Not the owner"
// @Failure      404     {object}  models.ErrorResponse "Post not found"
// @Router       /api/posts/{id}/content [put]
func (h *PostHandler) EditPostContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Check ownership - only post owner can edit
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		if err == repository.ErrPostNotFound {
			utils.ErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}
	if post.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only edit your own posts", http.StatusForbidden)
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

// DeletePost performs a soft-delete on a post
// @Summary      Delete post
// @Description  Marks a post as deleted without removing it from the database. Only the owner can delete it.
// @Tags         Posts
// @Security     CookieAuth
// @Param        id    path      string  true  "Post ID"
// @Success      200
// @Failure      403   {object}  models.ErrorResponse "Forbidden"
// @Failure      404   {object}  models.ErrorResponse "Post not found"
// @Router       /api/posts/{id} [delete]
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Check ownership - only post owner can delete
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		if err == repository.ErrPostNotFound {
			utils.ErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}
	if post.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only delete your own posts", http.StatusForbidden)
		return
	}

	if err := h.PostRepo.SoftDeletePost(postID); err != nil {
		utils.ErrorResponse(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
