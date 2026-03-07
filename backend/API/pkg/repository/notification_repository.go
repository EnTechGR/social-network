package repository

import (
	"database/sql"
	"time"

	"social-network/pkg/models"
	"social-network/pkg/utils"
)

// NotificationRepository handles CRUD operations for notifications
type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a new notification
func (r *NotificationRepository) Create(n *models.Notification) error {
	n.ID = utils.GenerateUUID()
	n.CreatedAt = time.Now()
	n.Visible = true

	var postID interface{}
	if n.PostID != "" {
		postID = n.PostID
	}

	var commentID interface{}
	if n.CommentID != nil && *n.CommentID != "" {
		commentID = *n.CommentID
	}

	_, err := r.db.Exec(`INSERT INTO notifications (notification_id, user_id, from_user_id, type, post_id, comment_id, created_at, is_read, is_visible) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.UserID, n.FromUserID, n.Type, postID, commentID, n.CreatedAt, n.Read, n.Visible)
	return err
}

// GetByUser returns notifications for a user ordered by newest first
// GetByUser returns notifications for a user ordered by newest first. It joins
// the user table to fetch the username of the user who triggered the
// notification (from_user_id).
func (r *NotificationRepository) GetByUser(userID string) ([]models.NotificationView, error) {
	rows, err := r.db.Query(`
                SELECT n.notification_id, COALESCE(n.from_user_id, ''), COALESCE(u.nickname, ''), n.type, COALESCE(n.post_id, ''), n.comment_id, n.created_at, n.is_read, n.is_visible
                FROM notifications n
                LEFT JOIN user u ON n.from_user_id = u.user_id
                WHERE n.user_id = ? AND n.is_visible = 1
                ORDER BY n.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.NotificationView
	for rows.Next() {
		var n models.NotificationView
		if err := rows.Scan(&n.ID, &n.FromUserID, &n.Nickname, &n.Type, &n.PostID, &n.CommentID, &n.CreatedAt, &n.Read, &n.Visible); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notifs, nil
}

// CountByUser returns the number of notifications for a user.
func (r *NotificationRepository) CountByUser(userID string) (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_visible = 1`, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Hide marks a notification as not visible for the given user.
func (r *NotificationRepository) Hide(id, userID string) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_visible = 0 WHERE notification_id = ? AND user_id = ?`, id, userID)
	return err
}
