package handlers

import (
	"log"

	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/websocket"
)

func createAndPushNotification(
	notificationRepo *repository.NotificationRepository,
	hub *websocket.Hub,
	targetUserID,
	fromUserID,
	fromNickname,
	notificationType string,
) {
	if notificationRepo == nil {
		return
	}

	n := models.Notification{
		UserID:     targetUserID,
		FromUserID: fromUserID,
		Type:       notificationType,
	}

	if err := notificationRepo.Create(&n); err != nil {
		log.Printf("[NotificationHelper] Failed to create %s notification for user %s: %v", notificationType, targetUserID, err)
		return
	}

	if hub == nil {
		return
	}

	hub.SendNotification(targetUserID, models.NotificationView{
		ID:         n.ID,
		FromUserID: fromUserID,
		Nickname:   fromNickname,
		Type:       notificationType,
		PostID:     "",
		CommentID:  nil,
		CreatedAt:  n.CreatedAt,
		Read:       false,
		Visible:    true,
	})
}
