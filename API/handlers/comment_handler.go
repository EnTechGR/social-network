package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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
		Hub:              hub,
	}
}


// CreateComment creates a new comment on a post
// @Summary      Create a comment
// @Description  Adds a new comment to a specific post. Triggers a real-time notification for the post owner via WebSocket.
// @Tags         Comments
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        comment  body      handlers.CreateCommentRequest  true  "Comment details"
// @Success      201      {object}  models.Comment
// @Failure      400      {object}  models.ErrorResponse "Missing PostID or Content"
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Failure      500      {object}  models.ErrorResponse "Database error"
// @Router       /api/comments [post]
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

	// ✅ Notify post owner about new comment
	if post, err := h.PostRepo.GetByID(req.PostID); err == nil && post != nil && post.UserID != user.ID {
		n := models.Notification{
			UserID:     post.UserID,
			FromUserID: user.ID,
			Type:       "comment",
			PostID:     req.PostID,
			CommentID:  &created.ID,
		}
		
		if err := h.NotificationRepo.Create(n); err != nil {
			log.Printf("[CommentHandler] Failed to create notification: %v", err)
		} else {
			// ✅ Send real-time notification via WebSocket
			if h.Hub != nil {
				notificationView := models.NotificationView{
					ID:        n.ID,
					Nickname:  user.Nickname,
					Type:      "comment",
					PostID:    req.PostID,
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

	utils.JSONResponse(w, created, http.StatusCreated)
}

// EditComment updates an existing comment
// @Summary      Edit a comment
// @Description  Updates the text content of an existing comment. Only the author can perform this action.
// @Tags         Comments
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Comment ID"
// @Param        content  body      handlers.EditCommentRequest   true  "New content"
// @Success      200      {object}  map[string]string             "status: updated"
// @Failure      403      {object}  models.ErrorResponse          "Forbidden - not the owner"
// @Failure      404      {object}  models.ErrorResponse          "Comment not found"
// @Router       /api/comments/{id} [put]
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
			
			if err := h.NotificationRepo.Create(n); err != nil {
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
// @Router       /api/comments/{id} [delete]
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
			
			if err := h.NotificationRepo.Create(n); err != nil {
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