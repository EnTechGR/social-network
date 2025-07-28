package repository

import (
	"database/sql"
	"time"

	"forum/models"
	"forum/utils"
)

// NotificationRepository handles CRUD operations for notifications
type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a new notification
func (r *NotificationRepository) Create(n models.Notification) error {
	n.ID = utils.GenerateUUID()
	n.CreatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO notifications (notification_id, user_id, from_user_id, type, post_id, comment_id, created_at, is_read) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.UserID, n.FromUserID, n.Type, n.PostID, n.CommentID, n.CreatedAt, n.Read)
	return err
}

// GetByUser returns notifications for a user ordered by newest first
// GetByUser returns notifications for a user ordered by newest first. It joins
// the user table to fetch the username of the user who triggered the
// notification (from_user_id).
func (r *NotificationRepository) GetByUser(userID string) ([]models.NotificationView, error) {
	rows, err := r.db.Query(`
                SELECT n.notification_id, u.username, n.type, n.post_id, n.comment_id, n.created_at, n.is_read
                FROM notifications n
                JOIN user u ON n.from_user_id = u.user_id
                WHERE n.user_id = ?
                ORDER BY n.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.NotificationView
	for rows.Next() {
		var n models.NotificationView
		if err := rows.Scan(&n.ID, &n.Username, &n.Type, &n.PostID, &n.CommentID, &n.CreatedAt, &n.Read); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// CountByUser returns the number of notifications for a user.
func (r *NotificationRepository) CountByUser(userID string) (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ?`, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
