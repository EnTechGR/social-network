package models

import "time"

// Comment represents a single comment on a post
// swagger:model Comment
type Comment struct {
	// The unique identifier for the comment (UUID)
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// The ID of the post this comment belongs to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// The ID of the user who authored the comment
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The text content of the comment
	// example: This is a very insightful comment!
	Content *string `json:"content"`
	// The timestamp when the comment was created
	// example: 2025-12-21T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the comment was last updated
	// example: 2025-12-21T12:30:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// CommentWithUser represents a comment with expanded author information
// swagger:model CommentWithUser
type CommentWithUser struct {
	// The unique identifier for the comment (UUID)
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// The ID of the post this comment belongs to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// The ID of the user who authored the comment
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The display name/nickname of the comment author
	// example: johndoe_99
	Nickname string `json:"nickname"`
	// The text content of the comment
	// example: This is a very insightful comment!
	Content *string `json:"content"`
	// The timestamp when the comment was created
	// example: 2025-12-21T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the comment was last updated
	// example: 2025-12-21T12:30:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}