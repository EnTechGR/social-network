package models

import "time"

// Notification represents a notification for a user about activity on their posts
// swagger:model Notification
type Notification struct {
	// The unique identifier for the notification (UUID)
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the user who receives the notification
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The ID of the user who triggered the notification
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	FromUserID string `json:"from_user_id"`
	// The type of activity.
	// Enum: like, dislike, comment, edit_comment, delete_comment
	// example: comment
	Type string `json:"type"`
	// The ID of the post the activity relates to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// The ID of the comment (if applicable). Nil for likes/dislikes on posts.
	// example: 550e8400-e29b-41d4-a716-446655440000
	CommentID *string `json:"comment_id,omitempty"`
	// The timestamp when the notification was generated
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// Whether the notification has been read by the user
	// example: false
	Read bool `json:"read"`
	// Whether the notification is currently visible in the UI
	// example: true
	Visible bool `json:"visible"`
}

// NotificationView represents a notification with the sender's nickname instead of ID
// swagger:model NotificationView
type NotificationView struct {
	// The unique identifier for the notification (UUID)
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the user who triggered the notification
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	FromUserID string `json:"from_user_id,omitempty"`
	// The nickname of the user who triggered the notification
	// example: johndoe_99
	Nickname string `json:"nickname"`
	// The type of activity.
	// Enum: like, dislike, comment, edit_comment, delete_comment
	// example: like
	Type string `json:"type"`
	// The ID of the post the activity relates to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// The ID of the comment (if applicable)
	// example: 550e8400-e29b-41d4-a716-446655440000
	CommentID *string `json:"comment_id,omitempty"`
	// The timestamp when the notification was generated
	CreatedAt time.Time `json:"created_at"`
	// Whether the notification has been read
	Read bool `json:"read"`
	// Whether the notification is visible
	Visible bool `json:"visible"`
}
