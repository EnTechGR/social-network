package handlers

import (
	"fmt"
	"log"
	"net/http"

	"social-network/middleware"
	"social-network/repository"
	"social-network/utils"
)

// ============================================================================
// ImageHTTPHandler - HTTP wrapper around ImageRepository
// ============================================================================
//
// This handler provides HTTP endpoints for image operations and delegates
// the actual image processing to ImageRepository. It handles:
// - HTTP request parsing (multipart forms)
// - User authentication validation
// - Authorization checks (ownership)
// - HTTP response formatting
// - Error handling and appropriate status codes

type ImageHTTPHandler struct {
	imageRepo *repository.ImageRepository
	postRepo  *repository.PostRepository
}

// NewImageHTTPHandler creates a new HTTP handler for image operations
func NewImageHTTPHandler(imageRepo *repository.ImageRepository, postRepo *repository.PostRepository) *ImageHTTPHandler {
	return &ImageHTTPHandler{
		imageRepo: imageRepo,
		postRepo:  postRepo,
	}
}

// UploadAvatar uploads a user avatar image
// @Summary      Upload user avatar
// @Description  Uploads a new avatar image for the authenticated user. Replaces any existing avatar. Max size 5MB.
// @Tags         Images
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar  formData  file  true  "Avatar image file (JPEG, PNG, GIF)"
// @Success      200     {object}  map[string]interface{} "status: success, message: Avatar uploaded successfully"
// @Failure      400     {object}  models.ErrorResponse   "Invalid file or too large"
// @Failure      401     {object}  models.ErrorResponse   "Unauthorized"
// @Router       /api/v1/user/avatar [post]
func (h *ImageHTTPHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Authentication - Get current user from context
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse multipart form (max 5MB for avatars)
	const maxAvatarSize = 5 << 20 // 5MB
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		utils.ErrorResponse(w, "Invalid form data or file too large", http.StatusBadRequest)
		return
	}

	// 3. Extract file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(w, "Avatar file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Validate file size
	if header.Size > maxAvatarSize {
		utils.ErrorResponse(w, "Avatar exceeds 5 MB limit", http.StatusBadRequest)
		return
	}

	// 5. Call repository to process and save avatar
	if err := h.imageRepo.UploadUserAvatar(file, header, user.ID); err != nil {
		log.Printf("Failed to upload avatar for user %s: %v", user.ID, err)
		utils.ErrorResponse(w, fmt.Sprintf("Failed to upload avatar: %v", err), http.StatusInternalServerError)
		return
	}

	// 6. Success response
	utils.JSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": "Avatar uploaded successfully",
	}, http.StatusOK)
}

// UploadPostImages uploads images for a specific post
// @Summary      Upload post images
// @Description  Uploads one or more images and associates them with a post. Max total size 20MB.
// @Tags         Images
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        post_id  formData  string  true  "ID of the post"
// @Success      200      {object}  ImageUploadResponse
// @Failure      403      {object}  models.ErrorResponse
// @Router       /api/v1/images/upload [post]
func (h *ImageHTTPHandler) UploadPostImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Authentication
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse multipart form (max 20MB total for post images)
	const maxPostImagesSize = 20 << 20 // 20MB
	if err := r.ParseMultipartForm(maxPostImagesSize); err != nil {
		utils.ErrorResponse(w, "Invalid form data or files too large", http.StatusBadRequest)
		return
	}

	// 3. Extract post_id from form
	postID := r.FormValue("post_id")
	if postID == "" {
		utils.ErrorResponse(w, "Post ID required", http.StatusBadRequest)
		return
	}

	// 4. Authorization - Check ownership (only post owner can upload images)
	post, err := h.postRepo.GetPostByID(postID)
	if err != nil {
		if err == repository.ErrPostNotFound {
			utils.ErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to retrieve post %s: %v", postID, err)
			utils.ErrorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	if post.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only upload images to your own posts", http.StatusForbidden)
		return
	}

	// 5. Extract image files from form
	// The form field name should be "image" or "images"
	form := r.MultipartForm
	files := form.File["image"]
	if len(files) == 0 {
		// Try alternative field name "images"
		files = form.File["images"]
	}

	if len(files) == 0 {
		utils.ErrorResponse(w, "At least one image file required", http.StatusBadRequest)
		return
	}

	// 6. Call repository to process and save images
	if err := h.imageRepo.UploadPostImages(files, postID, user.ID); err != nil {
		log.Printf("Failed to upload post images for post %s: %v", postID, err)
		utils.ErrorResponse(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
		return
	}

	// 7. Success response
	utils.JSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("%d image(s) uploaded successfully", len(files)),
		"count":   len(files),
	}, http.StatusOK)
}

// DeletePostImages deletes all images associated with a post
// @Summary      Delete post images
// @Description  Removes all image associations for a post. Only the post owner can perform this.
// @Tags         Images
// @Security     CookieAuth
// @Produce      json
// @Param        postId   path      string  true  "Post ID"
// @Success      200      {object}  SimpleSuccessResponse
// @Failure      401      {object}  models.ErrorResponse
// @Router       /api/v1/images/delete/{postId} [delete]
func (h *ImageHTTPHandler) DeletePostImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Authentication
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Extract postID from URL path
	postID := utils.GetLastPathParam(r)
	if postID == "" {
		utils.ErrorResponse(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// 3. Authorization - Check ownership
	post, err := h.postRepo.GetPostByID(postID)
	if err != nil {
		if err == repository.ErrPostNotFound {
			utils.ErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to retrieve post %s: %v", postID, err)
			utils.ErrorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	if post.UserID != user.ID {
		utils.ErrorResponse(w, "Forbidden - you can only delete images from your own posts", http.StatusForbidden)
		return
	}

	// 4. Call repository to delete images (soft delete)
	if err := h.imageRepo.DeletePostImages(postID); err != nil {
		log.Printf("Failed to delete images for post %s: %v", postID, err)
		utils.ErrorResponse(w, "Failed to delete images", http.StatusInternalServerError)
		return
	}

	// 5. Success response
	utils.JSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": "Post images deleted successfully",
	}, http.StatusOK)
}