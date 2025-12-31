package models

import "time"

// Post represents a social-network post in its raw database format
// swagger:model Post
type Post struct {
	// The unique identifier for the post (UUID)
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	ID string `json:"id"`
	// The ID of the user who created the post
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// A list of category IDs associated with this post
	// example: [1, 2, 5]
	CategoryIDs []int `json:"category_id"`
	// The title of the post
	// example: How to use Go with Docker?
	Title *string `json:"title"`
	// The main body content of the post
	// example: I am trying to containerize my Go application...
	Content *string `json:"content"`
	// The timestamp when the post was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the post was last updated
	// example: 2025-12-29T19:30:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// PostWithUser represents a post with author nickname and image URLs for the UI
// swagger:model PostWithUser
type PostWithUser struct {
	// The unique identifier for the post (UUID)
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	ID string `json:"id"`
	// The ID of the author
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The display name of the author
	// example: gopher_expert
	Nickname string `json:"nickname"`
	// The primary category ID for this post
	// example: 1
	CategoryID int `json:"category_id"`
	// The title of the post
	Title *string `json:"title"`
	// The body content of the post
	Content *string `json:"content"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: 2025-12-29T19:30:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// URL to the full-size image attached to the post
	// example: /uploads/images/posts/7b1a2c3d.jpg
	ImageURL string `json:"image_url,omitempty"`
	// URL to the thumbnail version of the image
	// example: /uploads/images/thumbnails/7b1a2c3d.jpg
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}