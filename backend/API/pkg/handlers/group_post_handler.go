package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"social-network/pkg/middleware"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"
)

// GroupPostHandler handles group post related endpoints
type GroupPostHandler struct {
	PostRepo  *repository.PostRepository
	ImageRepo *repository.ImageRepository
}

// NewGroupPostHandler creates a new GroupPostHandler
func NewGroupPostHandler(postRepo *repository.PostRepository, imageRepo *repository.ImageRepository) *GroupPostHandler {
	return &GroupPostHandler{
		PostRepo:  postRepo,
		ImageRepo: imageRepo,
	}
}

// CreateGroupPost creates a new post within a group
// @Summary      Create a group post
// @Description  Creates a new post within a group. Only group members can create posts.
// @Tags         Group Posts
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id      path      string  true   "Group ID"
// @Param        title   formData  string  true   "Post title (max 200 chars)"
// @Param        content formData  string  true   "Post content (max 2000 chars)"
// @Param        image   formData  file    false  "Images to attach (multiple allowed, max 20MB total)"
// @Success      201   {object}  map[string]interface{} "Post created successfully"
// @Failure      400   {object}  models.ErrorResponse "Invalid request"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not a group member"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/groups/{id}/posts/create [post]
func (h *GroupPostHandler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	groupID := utils.GetLastPathParam(r)
	if groupID == "" {
		utils.ErrorResponse(w, "Missing group ID", http.StatusBadRequest)
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

	// Create the group post
	post := models.Post{
		UserID:     user.ID,
		GroupID:    &groupID,
		Visibility: "public", // Group posts are public within the group
		Title:      &title,
		Content:    &content,
	}

	created, err := h.PostRepo.CreateGroupPost(post, user.ID)
	if err != nil {
		log.Printf("ERROR creating group post: %v", err)
		utils.ErrorResponse(w, "Failed to create group post. You may not be a member of this group.", http.StatusForbidden)
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
		"message":         "Group post created successfully",
	}

	if uploadedCount > 0 {
		response["message"] = fmt.Sprintf("Group post created with %d image(s)", uploadedCount)
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

// GetGroupPosts retrieves all posts for a specific group
// @Summary      Get group posts
// @Description  Retrieves all posts within a group. Only group members can view posts.
// @Tags         Group Posts
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.PostWithUser "List of group posts"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      403  {object}  models.ErrorResponse "Not a group member"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/groups/{id}/posts [get]
func (h *GroupPostHandler) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	groupID := utils.GetLastPathParam(r)
	if groupID == "" {
		utils.ErrorResponse(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	// Get group posts (repository will check membership)
	posts, err := h.PostRepo.GetGroupPosts(groupID, user.ID)
	if err != nil {
		log.Printf("ERROR getting group posts: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve group posts", http.StatusInternalServerError)
		return
	}

	// If posts is empty, it could mean:
	// 1. No posts in the group yet
	// 2. User is not a member
	// We'll return empty array in both cases for simplicity
	utils.JSONResponse(w, posts, http.StatusOK)
}

// GetGroupPostByID retrieves a specific group post
// @Summary      Get a specific group post
// @Description  Retrieves details of a specific group post. Only group members can view.
// @Tags         Group Posts
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Post ID"
// @Success      200  {object}  models.PostWithUser "Post details"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      403  {object}  models.ErrorResponse "Not a group member"
// @Failure      404  {object}  models.ErrorResponse "Post not found"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/groups/posts/{id} [get]
func (h *GroupPostHandler) GetGroupPostByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract post ID from URL path
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Get group post (repository will check membership)
	post, err := h.PostRepo.GetGroupPostByID(postID, user.ID)
	if err != nil {
		log.Printf("ERROR getting group post: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve group post", http.StatusInternalServerError)
		return
	}

	if post == nil {
		utils.ErrorResponse(w, "Post not found or you are not a member of this group", http.StatusNotFound)
		return
	}

	utils.JSONResponse(w, post, http.StatusOK)
}