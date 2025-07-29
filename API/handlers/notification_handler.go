package handlers

import (
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// NotificationHandler handles fetching notifications for a user
type NotificationHandler struct {
	Repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{Repo: repo}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

// HideNotification marks a notification as not visible for the current user.
func (h *NotificationHandler) HideNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
