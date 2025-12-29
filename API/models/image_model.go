package models

import "time"

// Image represents an uploaded image associated with a post
// swagger:model Image
type Image struct {
	// The unique identifier for the image (UUID)
	// example: d290f1ee-6c54-4b01-90e6-d701748f0851
	ID string `json:"id"`
	// The ID of the post this image belongs to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// The ID of the user who uploaded the image
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The server-side path to the full-sized image
	// example: /uploads/images/2025/12/nature-shot.jpg
	FilePath string `json:"file_path"`
	// The server-side path to the generated thumbnail
	// example: /uploads/images/thumbnails/nature-shot.jpg
	ThumbnailPath string `json:"thumbnail_path"`
	// The timestamp when the image was uploaded
	// example: 2025-12-21T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
}