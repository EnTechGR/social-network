package models

import "time"

type Post struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	CategoryIDs []int      `json:"category_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// PostWithUser is a post along with the username of its author
type PostWithUser struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	CategoryID   int       `json:"category_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	ImageURL     string    `json:"image_url,omitempty"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
}
