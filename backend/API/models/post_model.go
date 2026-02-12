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
	// The ID of the group this post belongs to (null for non-group posts)
	// example: g1h2i3j4-k5l6-m7n8-o9p0-q1r2s3t4u5v6
	GroupID *string `json:"group_id,omitempty"`
	// The visibility of the post: public, followers, or private
	// example: public
	Visibility string `json:"visibility"`
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
	// The ID of the group this post belongs to (null for non-group posts)
	// example: g1h2i3j4-k5l6-m7n8-o9p0-q1r2s3t4u5v6
	GroupID *string `json:"group_id,omitempty"`
	// The name of the group (if this is a group post)
	// example: Go Programming Group
	GroupName *string `json:"group_name,omitempty"`
	// The visibility of the post: public, followers, or private
	// example: public
	Visibility string `json:"visibility"`
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