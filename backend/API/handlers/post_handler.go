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

func parseAllowedUserIDs(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var ids []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &ids); err != nil {
			return nil, err
		}
	} else {
		ids = strings.Split(raw, ",")
	}

	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}

	return unique, nil
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
// @Param        visibility  formData  string  false  "Post visibility: public (default), followers, or private"
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

	var title, content, visibility string
	var allowedUserIDs []string

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
		visibility = r.FormValue("visibility")
		if visibility == "" {
			visibility = "public" // Default to public
		}

		rawAllowedValues := r.MultipartForm.Value["allowed_user_ids"]
		for _, raw := range rawAllowedValues {
			parsedIDs, err := parseAllowedUserIDs(raw)
			if err != nil {
				utils.ErrorResponse(w, "Invalid allowed_user_ids format", http.StatusBadRequest)
				return
			}
			allowedUserIDs = append(allowedUserIDs, parsedIDs...)
		}

		// Support singular field name as fallback.
		if len(allowedUserIDs) == 0 {
			parsedIDs, err := parseAllowedUserIDs(r.FormValue("allowed_user_id"))
			if err != nil {
				utils.ErrorResponse(w, "Invalid allowed_user_id format", http.StatusBadRequest)
				return
			}
			allowedUserIDs = append(allowedUserIDs, parsedIDs...)
		}

		log.Printf("DEBUG: Extracted from multipart - title: %s, content: %s, visibility: %s", title, content, visibility)
	} else {
		// Parse as JSON (backward compatibility)
		log.Printf("DEBUG: Attempting to parse as JSON")
		var req struct {
			Title          string   `json:"title"`
			Content        string   `json:"content"`
			Visibility     string   `json:"visibility"`
			AllowedUserIDs []string `json:"allowed_user_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR parsing JSON: %v", err)
			utils.ErrorResponse(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		title = req.Title
		content = req.Content
		visibility = req.Visibility
		parsedIDs := make([]string, 0, len(req.AllowedUserIDs))
		for _, raw := range req.AllowedUserIDs {
			ids, err := parseAllowedUserIDs(raw)
			if err != nil {
				utils.ErrorResponse(w, "Invalid allowed_user_ids format", http.StatusBadRequest)
				return
			}
			parsedIDs = append(parsedIDs, ids...)
		}
		allowedUserIDs = parsedIDs
		if visibility == "" {
			visibility = "public" // Default to public
		}
		log.Printf("DEBUG: Extracted from JSON - title: %s, content: %s, visibility: %s, allowed users: %d", title, content, visibility, len(allowedUserIDs))
	}

	// Validate required fields
	if title == "" || content == "" {
		utils.ErrorResponse(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Validate visibility
	if visibility != "public" && visibility != "followers" && visibility != "private" {
		utils.ErrorResponse(w, "Invalid visibility. Must be: public, followers, or private", http.StatusBadRequest)
		return
	}

	// Create the post
	post := models.Post{
		UserID:     user.ID,
		Visibility: visibility,
		Title:      &title,
		Content:    &content,
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

	// Handle allowed users for private posts
	var allowedCount int
	if visibility == "private" && len(allowedUserIDs) > 0 {
		for _, userID := range allowedUserIDs {
			if err := h.PostRepo.AddAllowedUser(created.ID, userID); err != nil {
				log.Printf("Warning: Failed to add allowed user %s to post %s: %v", userID, created.ID, err)
				// Don't fail post creation if adding allowed users fails, just log it
			} else {
				allowedCount++
			}
		}
	}

	// Return success with upload count
	response := map[string]interface{}{
		"post":            created,
		"images_uploaded": uploadedCount,
	}

	if allowedCount > 0 {
		response["allowed_users_added"] = allowedCount
	}

	if uploadedCount > 0 {
		response["message"] = fmt.Sprintf("Post created with %d image(s)", uploadedCount)
		if allowedCount > 0 {
			response["message"] = fmt.Sprintf("Post created with %d image(s) and %d allowed user(s)", uploadedCount, allowedCount)
		}
	} else if allowedCount > 0 {
		response["message"] = fmt.Sprintf("Post created with %d allowed user(s)", allowedCount)
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

// GetUserPosts retrieves posts from a specific user, respecting visibility settings
// @Summary      Get user's posts
// @Description  Retrieves posts from a specific user based on visibility rules
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "User ID"
// @Success      200   {array}   models.PostWithUser "List of visible posts"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      404   {object}  models.ErrorResponse "User not found"
// @Router       /api/v1/users/{id}/posts [get]
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := middleware.GetCurrentUser(r)
	if currentUser == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetLastPathParam(r)
	if userID == "" {
		utils.ErrorResponse(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get all posts from the target user
	allPosts, err := h.PostRepo.GetPostsByUser(userID)
	if err != nil {
		log.Printf("Error fetching posts for user %s: %v", userID, err)
		utils.ErrorResponse(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts based on visibility
	var visiblePosts []models.PostWithUser
	for _, post := range allPosts {
		// Owner can always see their own posts
		if post.UserID == currentUser.ID {
			visiblePosts = append(visiblePosts, post)
			continue
		}

		// Check visibility
		switch post.Visibility {
		case "public":
			// Anyone can see public posts
			visiblePosts = append(visiblePosts, post)

		case "followers":
			// Only followers of the post owner can see it
			isFollower, err := h.PostRepo.CheckIfFollower(currentUser.ID, userID)
			if err == nil && isFollower {
				visiblePosts = append(visiblePosts, post)
			}

		case "private":
			// Check if current user is in the allowed users list
			allowedUsers, err := h.PostRepo.GetAllowedUsers(post.ID)
			if err == nil {
				for _, allowedID := range allowedUsers {
					if allowedID == currentUser.ID {
						visiblePosts = append(visiblePosts, post)
						break
					}
				}
			}
		}
	}

	utils.JSONResponse(w, visiblePosts, http.StatusOK)
}

// UpdatePostVisibility updates the visibility of a specific post
// @Summary      Update post visibility
// @Description  Allows the post owner to change the visibility of their post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        data  body      object  true  "New visibility {\"visibility\": \"public|followers|private\"}"
// @Success      200   {object}  map[string]string "status: visibility updated"
// @Failure      400   {object}  models.ErrorResponse "Invalid request or invalid visibility"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not the post owner"
// @Failure      404   {object}  models.ErrorResponse "Post not found"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/update-visibility/{id} [put]
func (h *PostHandler) UpdatePostVisibility(w http.ResponseWriter, r *http.Request) {
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
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Visibility == "" {
		utils.ErrorResponse(w, "Visibility is required", http.StatusBadRequest)
		return
	}

	// Validate visibility
	if req.Visibility != "public" && req.Visibility != "followers" && req.Visibility != "private" {
		utils.ErrorResponse(w, "Invalid visibility. Must be: public, followers, or private", http.StatusBadRequest)
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

	// Update visibility
	if err := h.PostRepo.UpdateVisibility(postID, req.Visibility); err != nil {
		utils.ErrorResponse(w, "Failed to update visibility", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "visibility updated"}, http.StatusOK)
}

// AddAllowedUser adds a specific user to the allowed viewers of a private post
// @Summary      Add allowed user to post
// @Description  Allows the post owner to grant access to a specific user for a private post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        data  body      object  true  "User to allow {\"user_id\": \"uuid\"}"
// @Success      201   {object}  map[string]string "status: user added"
// @Failure      400   {object}  models.ErrorResponse "Invalid request"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not the post owner"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/{id}/add-allowed-user [post]
func (h *PostHandler) AddAllowedUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		utils.ErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.PostRepo.CheckOwnership(postID, user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to verify ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		utils.ErrorResponse(w, "Not authorized to modify this post", http.StatusForbidden)
		return
	}

	// Add allowed user
	if err := h.PostRepo.AddAllowedUser(postID, req.UserID); err != nil {
		utils.ErrorResponse(w, "Failed to add allowed user", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "user added"}, http.StatusCreated)
}

// RemoveAllowedUser removes a user from the allowed viewers of a private post
// @Summary      Remove allowed user from post
// @Description  Allows the post owner to revoke access to a specific user for a private post.
// @Tags         Posts
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Post ID"
// @Param        data  body      object  true  "User to remove {\"user_id\": \"uuid\"}"
// @Success      200   {object}  map[string]string "status: user removed"
// @Failure      400   {object}  models.ErrorResponse "Invalid request"
// @Failure      401   {object}  models.ErrorResponse "Unauthorized"
// @Failure      403   {object}  models.ErrorResponse "Not the post owner"
// @Failure      500   {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/posts/{id}/remove-allowed-user [post]
func (h *PostHandler) RemoveAllowedUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		utils.ErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check ownership
	isOwner, err := h.PostRepo.CheckOwnership(postID, user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to verify ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		utils.ErrorResponse(w, "Not authorized to modify this post", http.StatusForbidden)
		return
	}

	// Remove allowed user
	if err := h.PostRepo.RemoveAllowedUser(postID, req.UserID); err != nil {
		utils.ErrorResponse(w, "Failed to remove allowed user", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "user removed"}, http.StatusOK)
}
