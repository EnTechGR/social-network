package handlers

import (
	"net/http"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/utils"
	"social-network/websocket"
)

type NotificationHandler struct {
	Repo *repository.NotificationRepository
	Hub  *websocket.Hub // ✅ Add this
}

func NewNotificationHandler(
	repo *repository.NotificationRepository,
	hub *websocket.Hub, // ✅ Add this parameter
) *NotificationHandler {
	return &NotificationHandler{
		Repo: repo,
		Hub:  hub,
	}
}

// GetNotifications retrieves all active notifications for the user
// @Summary      Get notifications
// @Description  Returns a list of notifications for the authenticated user, including the total count of active notifications.
// @Tags         Notifications
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Failure      500      {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/notifications [get]
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	notifs, err := h.Repo.GetByUser(user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}

	count, err := h.Repo.CountByUser(user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Count         int                       `json:"count"`
		Notifications []models.NotificationView `json:"notifications"`
	}{
		Count:         count,
		Notifications: notifs,
	}

	utils.JSONResponse(w, resp, http.StatusOK)
}

// HideNotification marks a notification as hidden or deleted
// @Summary      Hide notification
// @Description  Soft-deletes or hides a specific notification from the user's view. Requires the notification ID as a path parameter.
// @Tags         Notifications
// @Security     CookieAuth
// @Param        id       path      string  true  "Notification ID"
// @Success      200
// @Failure      400      {object}  models.ErrorResponse "Missing notification ID"
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Router       /api/notifications/{id} [delete]
func (h *NotificationHandler) HideNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	notifID := utils.GetLastPathParam(r)
	if notifID == "" {
		utils.ErrorResponse(w, "Missing notification ID", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Hide(notifID, user.ID); err != nil {
		utils.ErrorResponse(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}
