package repository

import (
	"database/sql"
	"os"
	"time"

	"forum/models"
	"forum/utils"
)

type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(img models.Image) (*models.Image, error) {
	img.ID = utils.GenerateUUID()
	img.CreatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO images (image_id, post_id, user_id, file_path, thumbnail_path, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		img.ID, img.PostID, img.UserID, img.FilePath, img.ThumbnailPath, img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *ImageRepository) GetByPostID(postID string) ([]models.Image, error) {
	rows, err := r.db.Query(`SELECT image_id, post_id, user_id, file_path, thumbnail_path, created_at FROM images WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.PostID, &img.UserID, &img.FilePath, &img.ThumbnailPath, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

// DeleteByPostID deletes all images for a post from DB and filesystem
func (r *ImageRepository) DeleteByPostID(postID string) error {
	images, err := r.GetByPostID(postID)
	if err != nil {
		return err
	}
	for _, img := range images {
		_ = os.Remove(img.FilePath)
		_ = os.Remove(img.ThumbnailPath)
	}
	_, err = r.db.Exec(`DELETE FROM images WHERE post_id = ?`, postID)
	return err
}
