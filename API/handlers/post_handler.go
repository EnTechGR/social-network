package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/utils"
)

// PostHandler handles post related endpoints
type PostHandler struct {
	PostRepo  *repository.PostRepository
	ImageRepo *repository.ImageRepository
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(repo *repository.PostRepository, imageRepo *repository.ImageRepository) *PostHandler {
	return &PostHandler{PostRepo: repo, ImageRepo: imageRepo}
}

// CreatePost creates a new social-network post with optional image uploads
// @Summary      Create a post
// @Description  Creates a new post for the authenticated user with title and content. Optionally upload images in the same request.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        title  formData  string  true   "Post title (max 200 chars)"
// @Param        content  formData  string  true   "Post content (max 2000 chars)"
// @Param        image  formData  file    false  "Images to attach (multiple allowed, max 20MB total)"
// @Success      201   {object}  map[string]interface{} "Post with optional images uploaded count"
// @Failure      400   {object}  models.ErrorResponse "Invalid request body or missing fields"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/create [post]
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

	var title, content string

	// Determine content type and parse accordingly
	contentType := r.Header.Get("Content-Type")
	log.Printf("DEBUG: Content-Type header: %s", contentType)

	// Check if it's multipart (includes boundary info)
	if strings.Contains(contentType, "multipart/form-data") {
		// Parse as multipart/form-data
		const maxPostSize = 20 << 20 // 20MB
		log.Printf("DEBUG: Attempting to parse as multipart/form-data")
		if err := r.ParseMultipartForm(maxPostSize); err != nil {
			log.Printf("ERROR parsing multipart form: %v", err)
			utils.ErrorResponse(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
			return
		}

		title = r.FormValue("title")
		content = r.FormValue("content")
		log.Printf("DEBUG: Extracted from multipart - title: %s, content: %s", title, content)
	} else {
		// Parse as JSON (backward compatibility)
		log.Printf("DEBUG: Attempting to parse as JSON")
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR parsing JSON: %v", err)
			utils.ErrorResponse(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		title = req.Title
		content = req.Content
		log.Printf("DEBUG: Extracted from JSON - title: %s, content: %s", title, content)
	}

	// Validate required fields
	if title == "" || content == "" {
		utils.ErrorResponse(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create the post
	post := models.Post{
		UserID:  user.ID,
		Title:   &title,
		Content: &content,
	}

	created, err := h.PostRepo.Create(post)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Handle optional image uploads (only for multipart requests)
	var uploadedCount int
	if strings.Contains(contentType, "multipart/form-data") && r.MultipartForm != nil {
		form := r.MultipartForm
		files := form.File["image"]
		if len(files) == 0 {
			// Try alternative field name
			files = form.File["images"]
		}

		if len(files) > 0 {
			if err := h.ImageRepo.UploadPostImages(files, created.ID, user.ID); err != nil {
				log.Printf("Warning: Failed to upload images for post %s: %v", created.ID, err)
				// Don't fail the post creation if image upload fails
				// Return success with a warning message
				utils.JSONResponse(w, map[string]interface{}{
					"post":            created,
					"images_uploaded": 0,
					"warning":         fmt.Sprintf("Post created but image upload failed: %v", err),
				}, http.StatusCreated)
				return
			}
			uploadedCount = len(files)
		}
	}

	// Return success with upload count
	response := map[string]interface{}{
		"post":            created,
		"images_uploaded": uploadedCount,
	}

	if uploadedCount > 0 {
		response["message"] = fmt.Sprintf("Post created with %d image(s)", uploadedCount)
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

// EditPostTitle updates the title of a specific post
// @Summary      Edit post title
// @Description  Allows the post owner to update only the title of their post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        data  body      object  true  "New title {\"title\": \"Updated Title\"}"
// @Success      200   {object}  map[string]string "status: title updated"
// @Failure      400   {object}  models.ErrorResponse "Invalid request or missing title"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not the post owner"
// @Failure      404   {object}  models.ErrorResponse "Post not found"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/edit-title/{id} [put]
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

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		utils.ErrorResponse(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.PostRepo.CheckOwnership(postID, user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to verify ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		utils.ErrorResponse(w, "Not authorized to edit this post", http.StatusForbidden)
		return
	}

	// Update title
	if err := h.PostRepo.UpdateTitle(postID, req.Title); err != nil {
		utils.ErrorResponse(w, "Failed to update title", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "title updated"}, http.StatusOK)
}

// EditPostContent updates the content of a specific post
// @Summary      Edit post content
// @Description  Allows the post owner to update only the main text content of their post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        data  body      object  true  "New content {\"content\": \"Updated content\"}"
// @Success      200   {object}  map[string]string "status: content updated"
// @Failure      400   {object}  models.ErrorResponse "Invalid request or missing content"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not the post owner"
// @Failure      404   {object}  models.ErrorResponse "Post not found"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/edit-content/{id} [put]
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

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		utils.ErrorResponse(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.PostRepo.CheckOwnership(postID, user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to verify ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		utils.ErrorResponse(w, "Not authorized to edit this post", http.StatusForbidden)
		return
	}

	// Update content
	if err := h.PostRepo.UpdateContent(postID, req.Content); err != nil {
		utils.ErrorResponse(w, "Failed to update content", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "content updated"}, http.StatusOK)
}

// DeletePost marks a post as deleted
// @Summary      Delete post
// @Description  Marks a post as deleted without removing it from the database. Only the owner can delete it.
// @Tags         Posts
// @Security     CookieAuth
// @Param        id   path  string  true  "Post ID"
// @Success      200  {object}  map[string]string "status: deleted"
// @Failure      403  {object}  models.ErrorResponse "Forbidden"
// @Failure      404  {object}  models.ErrorResponse "Post not found"
// @Router       /api/v1/posts/delete/{id} [delete]
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

	// Check ownership
	isOwner, err := h.PostRepo.CheckOwnership(postID, user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to verify ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		utils.ErrorResponse(w, "Not authorized to delete this post", http.StatusForbidden)
		return
	}

	// Delete post
	if err := h.PostRepo.DeleteByID(postID); err != nil {
		utils.ErrorResponse(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
