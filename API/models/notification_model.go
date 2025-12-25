package models

import "time"

// Notification represents a notification for a user about activity on their posts
// Type can be: like, dislike, comment, edit_comment, delete_comment
// CommentID may be nil for like/dislike on posts
// For comment-related notifications, PostID is the post the comment belongs to.
type Notification struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FromUserID string    `json:"from_user_id"`
	Type       string    `json:"type"`
	PostID     string    `json:"post_id"`
	CommentID  *string   `json:"comment_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	Read       bool      `json:"read"`
	Visible    bool      `json:"visible"`
}

// NotificationView represents a notification with the sender's nickname
// instead of their user ID. This is used when returning notifications to the
// client.
type NotificationView struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Type      string    `json:"type"`
	PostID    string    `json:"post_id"`
	CommentID *string   `json:"comment_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
	Visible   bool      `json:"visible"`
}
