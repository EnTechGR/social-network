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
	"social-network/websocket"
)

// CommentHandler handles comment related endpoints
type CommentHandler struct {
	CommentRepo      *repository.CommentRepository
	PostRepo         *repository.PostRepository
	NotificationRepo *repository.NotificationRepository
	ImageRepo        *repository.ImageRepository
	Hub              *websocket.Hub // ✅ Add Hub reference
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(
	commentRepo *repository.CommentRepository,
	postRepo *repository.PostRepository,
	notifRepo *repository.NotificationRepository,
	hub *websocket.Hub, // ✅ Add Hub parameter
) *CommentHandler {
	return &CommentHandler{
		CommentRepo:      commentRepo,
		PostRepo:         postRepo,
		NotificationRepo: notifRepo,
		ImageRepo:        nil, // Will be set later if needed
		Hub:              hub,
	}
}

// SetImageRepo sets the image repository for comment image uploads
func (h *CommentHandler) SetImageRepo(imageRepo *repository.ImageRepository) {
	h.ImageRepo = imageRepo
}

// CreateComment creates a new comment on a post with optional image uploads
// @Summary      Create a comment
// @Description  Adds a new comment to a specific post. Optionally upload images. Triggers a real-time notification for the post owner via WebSocket.
// @Tags         Comments
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        post_id  formData  string  true   "Post ID"
// @Param        content  formData  string  true   "Comment content"
// @Param        image    formData  file    false  "Images to attach (multiple allowed)"
// @Success      201      {object}  map[string]interface{} "Comment with images_uploaded count"
// @Failure      400      {object}  models.ErrorResponse "Missing PostID or Content"
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Failure      500      {object}  models.ErrorResponse "Database error"
// @Router       /api/v1/comments/create [post]
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var postID, content string
	contentType := r.Header.Get("Content-Type")

	// Check if it's multipart (includes boundary info)
	if strings.Contains(contentType, "multipart/form-data") {
		// Parse as multipart/form-data
		const maxCommentSize = 20 << 20 // 20MB
		log.Printf("DEBUG: Attempting to parse comment as multipart/form-data")
		if err := r.ParseMultipartForm(maxCommentSize); err != nil {
			log.Printf("ERROR parsing multipart form: %v", err)
			utils.ErrorResponse(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
			return
		}

		postID = r.FormValue("post_id")
		content = r.FormValue("content")
		log.Printf("DEBUG: Extracted from multipart - post_id: %s, content: %s", postID, content)
	} else {
		// Parse as JSON (backward compatibility)
		log.Printf("DEBUG: Attempting to parse comment as JSON")
		var req struct {
			PostID  string `json:"post_id"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		postID = req.PostID
		content = req.Content
		log.Printf("DEBUG: Extracted from JSON - post_id: %s, content: %s", postID, content)
	}

	if postID == "" || content == "" {
		utils.ErrorResponse(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostID:  postID,
		UserID:  user.ID,
		Content: &content,
	}

	created, err := h.CommentRepo.Create(comment)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Handle optional image uploads (only for multipart requests)
	var uploadedCount int
	if strings.Contains(contentType, "multipart/form-data") && r.MultipartForm != nil && h.ImageRepo != nil {
		form := r.MultipartForm
		files := form.File["image"]
		if len(files) == 0 {
			files = form.File["images"]
		}

		if len(files) > 0 {
			// Note: Using post ID for image storage since comments share the same image table
			// This associates images with the post but tracks them via comment context
			if err := h.ImageRepo.UploadPostImages(files, postID, user.ID); err != nil {
				log.Printf("Warning: Failed to upload images for comment %s: %v", created.ID, err)
				// Don't fail the comment creation if image upload fails
				utils.JSONResponse(w, map[string]interface{}{
					"comment":         created,
					"images_uploaded": 0,
					"warning":         fmt.Sprintf("Comment created but image upload failed: %v", err),
				}, http.StatusCreated)
				return
			}
			uploadedCount = len(files)
		}
	}

	// ✅ Notify post owner about new comment
	if post, err := h.PostRepo.GetByID(postID); err == nil && post != nil && post.UserID != user.ID {
		n := models.Notification{
			UserID:     post.UserID,
			FromUserID: user.ID,
			Type:       "comment",
			PostID:     postID,
			CommentID:  &created.ID,
		}

		if err := h.NotificationRepo.Create(&n); err != nil {
			log.Printf("[CommentHandler] Failed to create notification: %v", err)
		} else {
			// ✅ Send real-time notification via WebSocket
			if h.Hub != nil {
				notificationView := models.NotificationView{
					ID:        n.ID,
					Nickname:  user.Nickname,
					Type:      "comment",
					PostID:    postID,
					CommentID: &created.ID,
					CreatedAt: n.CreatedAt,
					Read:      false,
					Visible:   true,
				}

				log.Printf("[CommentHandler] Sending WebSocket notification to user %s", post.UserID)
				h.Hub.SendNotification(post.UserID, notificationView)
			}
		}
	}

	// Return success with upload count
	response := map[string]interface{}{
		"comment":         created,
		"images_uploaded": uploadedCount,
	}

	if uploadedCount > 0 {
		response["message"] = fmt.Sprintf("Comment created with %d image(s)", uploadedCount)
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

// EditComment updates an existing comment
// @Summary      Edit a comment
// @Description  Updates the text content of an existing comment. Only the author can perform this action.
// @Tags         Comments
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Comment ID"
// @Success      200      {object}  map[string]string             "status: updated"
// @Failure      403      {object}  models.ErrorResponse          "Forbidden - not the owner"
// @Failure      404      {object}  models.ErrorResponse          "Comment not found"
// @Router       /api/v1/comments/{id} [put]
func (h *CommentHandler) EditComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
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

	// ✅ Notify post owner about edited comment
	if comment, err := h.CommentRepo.GetByID(commentID); err == nil && comment != nil {
		if post, err2 := h.PostRepo.GetByID(comment.PostID); err2 == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       "edit_comment",
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}

			if err := h.NotificationRepo.Create(&n); err != nil {
				log.Printf("[CommentHandler] Failed to create edit notification: %v", err)
			} else {
				// ✅ Send real-time notification via WebSocket
				if h.Hub != nil {
					notificationView := models.NotificationView{
						ID:        n.ID,
						Nickname:  user.Nickname,
						Type:      "edit_comment",
						PostID:    comment.PostID,
						CommentID: &comment.ID,
						CreatedAt: n.CreatedAt,
						Read:      false,
						Visible:   true,
					}

					h.Hub.SendNotification(post.UserID, notificationView)
				}
			}
		}
	}

	utils.JSONResponse(w, map[string]string{"status": "updated"}, http.StatusOK)
}

// DeleteComment soft-deletes a comment
// @Summary      Delete a comment
// @Description  Performs a soft-delete on a comment. Only the author can perform this action.
// @Tags         Comments
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Comment ID"
// @Success      200  {object}  map[string]string "status: deleted"
// @Failure      403  {object}  models.ErrorResponse "Forbidden"
// @Failure      404  {object}  models.ErrorResponse "Comment not found"
// @Router       /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
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

	// Actually delete the comment
	if err := h.CommentRepo.SoftDeleteComment(commentID); err != nil {
		utils.ErrorResponse(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// ✅ Create notification for post owner about deleted comment
	if comment != nil {
		if post, err := h.PostRepo.GetByID(comment.PostID); err == nil && post != nil && post.UserID != user.ID {
			n := models.Notification{
				UserID:     post.UserID,
				FromUserID: user.ID,
				Type:       "delete_comment",
				PostID:     comment.PostID,
				CommentID:  &comment.ID,
			}

			if err := h.NotificationRepo.Create(&n); err != nil {
				log.Printf("[CommentHandler] Failed to create delete notification: %v", err)
			} else {
				// ✅ Send real-time notification via WebSocket
				if h.Hub != nil {
					notificationView := models.NotificationView{
						ID:        n.ID,
						Nickname:  user.Nickname,
						Type:      "delete_comment",
						PostID:    comment.PostID,
						CommentID: &comment.ID,
						CreatedAt: n.CreatedAt,
						Read:      false,
						Visible:   true,
					}

					h.Hub.SendNotification(post.UserID, notificationView)
				}
			}
		}
	}

	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
