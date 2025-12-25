package models

import "time"

type Comment struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	UserID    string     `json:"user_id"`
	Content   *string    `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}


// CommentWithUser is a comment along with the nickname of its author
type CommentWithUser struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	UserID    string     `json:"user_id"`
	Nickname  string     `json:"nickname"`
	Content   *string    `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}