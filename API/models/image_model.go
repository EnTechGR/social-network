package models

import "time"

type Image struct {
	ID            string    `json:"id"`
	PostID        string    `json:"post_id"`
	UserID        string    `json:"user_id"`
	FilePath      string    `json:"file_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
	CreatedAt     time.Time `json:"created_at"`
}
