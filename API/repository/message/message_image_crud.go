package message

import (
	"database/sql"
	"fmt"
	"time"

	"forum/models"
)

// CreateMessageAndImage creates a message and the associated chat image within a single database transaction.
func (r *MessageRepository) CreateMessageAndImage(msg *models.Message, img *models.ChatImage) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after Rollback
		}
	}()

	// 1. Insert Message Record
	msgQuery := `
		INSERT INTO messages (message_id, sender_id, receiver_id, content, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(msgQuery, msg.MessageID, msg.SenderID, msg.ReceiverID, msg.Content, msg.CreatedAt, msg.IsRead)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert message record: %w", err)
	}

	// 2. Insert ChatImage Record
	// This uses the width and height values passed from the handler (which should be > 0)
	imgQuery := `
		INSERT INTO chat_images (
			image_id, message_id, user_id, filename, original_filename,
			file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(imgQuery,
		img.ImageID, img.MessageID, img.UserID, img.Filename, img.OriginalFilename,
		img.FilePath, img.ThumbnailPath, img.FileSize, img.MimeType, img.Width, img.Height, img.UploadedAt,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create chat image (transaction failed): %w", err)
	}

	// 3. Commit Transaction
	return tx.Commit()
}

// CreateChatImage inserts a new chat image into the database
func (r *MessageRepository) CreateChatImage(img *models.ChatImage) error {
	query := `
		INSERT INTO chat_images (
			image_id, message_id, user_id, filename, original_filename,
			file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.DB.Exec(
		query,
		img.ImageID,
		img.MessageID,
		img.UserID,
		img.Filename,
		img.OriginalFilename,
		img.FilePath,
		img.ThumbnailPath,
		img.FileSize,
		img.MimeType,
		img.Width,
		img.Height,
		img.UploadedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create chat image: %w", err)
	}

	return nil
}

// ... (Rest of the GetChatImageByID, GetChatImagesByMessageID, etc., functions remain unchanged) ...

// GetChatImageByID retrieves a chat image by its ID
func (r *MessageRepository) GetChatImageByID(imageID string) (*models.ChatImage, error) {
	query := `
		SELECT image_id, message_id, user_id, filename, original_filename,
			   file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		FROM chat_images
		WHERE image_id = ?
	`

	img := &models.ChatImage{}
	err := r.DB.QueryRow(query, imageID).Scan(
		&img.ImageID,
		&img.MessageID,
		&img.UserID,
		&img.Filename,
		&img.OriginalFilename,
		&img.FilePath,
		&img.ThumbnailPath,
		&img.FileSize,
		&img.MimeType,
		&img.Width,
		&img.Height,
		&img.UploadedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat image not found")
		}
		return nil, fmt.Errorf("failed to get chat image: %w", err)
	}

	return img, nil
}

// GetChatImagesByMessageID retrieves all images for a specific message
func (r *MessageRepository) GetChatImagesByMessageID(messageID string) ([]*models.ChatImage, error) {
	query := `
		SELECT image_id, message_id, user_id, filename, original_filename,
			   file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		FROM chat_images
		WHERE message_id = ?
		ORDER BY uploaded_at ASC
	`

	rows, err := r.DB.Query(query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat images: %w", err)
	}
	defer rows.Close()

	var images []*models.ChatImage
	for rows.Next() {
		img := &models.ChatImage{}
		err := rows.Scan(
			&img.ImageID,
			&img.MessageID,
			&img.UserID,
			&img.Filename,
			&img.OriginalFilename,
			&img.FilePath,
			&img.ThumbnailPath,
			&img.FileSize,
			&img.MimeType,
			&img.Width,
			&img.Height,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat images: %w", err)
	}

	return images, nil
}

// GetChatImagesByUserID retrieves all images uploaded by a specific user
func (r *MessageRepository) GetChatImagesByUserID(userID string, limit, offset int) ([]*models.ChatImage, error) {
	query := `
		SELECT image_id, message_id, user_id, filename, original_filename,
			   file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		FROM chat_images
		WHERE user_id = ?
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat images by user: %w", err)
	}
	defer rows.Close()

	var images []*models.ChatImage
	for rows.Next() {
		img := &models.ChatImage{}
		err := rows.Scan(
			&img.ImageID,
			&img.MessageID,
			&img.UserID,
			&img.Filename,
			&img.OriginalFilename,
			&img.FilePath,
			&img.ThumbnailPath,
			&img.FileSize,
			&img.MimeType,
			&img.Width,
			&img.Height,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat images: %w", err)
	}

	return images, nil
}

// DeleteChatImage removes a chat image from the database
func (r *MessageRepository) DeleteChatImage(imageID string) error {
	query := `DELETE FROM chat_images WHERE image_id = ?`

	result, err := r.DB.Exec(query, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete chat image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat image not found")
	}

	return nil
}

// DeleteChatImagesByMessageID removes all images associated with a message
func (r *MessageRepository) DeleteChatImagesByMessageID(messageID string) error {
	query := `DELETE FROM chat_images WHERE message_id = ?`

	_, err := r.DB.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete chat images by message: %w", err)
	}

	return nil
}

// GetTotalImagesSizeByUser calculates the total size of images uploaded by a user
func (r *MessageRepository) GetTotalImagesSizeByUser(userID string) (int64, error) {
	query := `SELECT COALESCE(SUM(file_size), 0) FROM chat_images WHERE user_id = ?`

	var totalSize int64
	err := r.DB.QueryRow(query, userID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total images size: %w", err)
	}

	return totalSize, nil
}

// GetImageCountByUser returns the number of images uploaded by a user
func (r *MessageRepository) GetImageCountByUser(userID string) (int, error) {
	query := `SELECT COUNT(*) FROM chat_images WHERE user_id = ?`

	var count int
	err := r.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user images: %w", err)
	}

	return count, nil
}

// GetChatImagesInConversation retrieves all images shared between two users
func (r *MessageRepository) GetChatImagesInConversation(userID1, userID2 string, limit, offset int) ([]*models.ChatImage, error) {
	query := `
		SELECT ci.image_id, ci.message_id, ci.user_id, ci.filename, ci.original_filename,
			   ci.file_path, ci.thumbnail_path, ci.file_size, ci.mime_type, 
			   ci.width, ci.height, ci.uploaded_at
		FROM chat_images ci
		INNER JOIN messages m ON ci.message_id = m.message_id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY ci.uploaded_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, userID1, userID2, userID2, userID1, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversation images: %w", err)
	}
	defer rows.Close()

	var images []*models.ChatImage
	for rows.Next() {
		img := &models.ChatImage{}
		err := rows.Scan(
			&img.ImageID,
			&img.MessageID,
			&img.UserID,
			&img.Filename,
			&img.OriginalFilename,
			&img.FilePath,
			&img.ThumbnailPath,
			&img.FileSize,
			&img.MimeType,
			&img.Width,
			&img.Height,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversation images: %w", err)
	}

	return images, nil
}

// ValidateMimeType checks if the provided MIME type is supported
func (r *MessageRepository) ValidateMimeType(mimeType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[mimeType]
}

// GetImagesByDateRange retrieves images uploaded within a specific date range
func (r *MessageRepository) GetImagesByDateRange(userID string, startDate, endDate time.Time) ([]*models.ChatImage, error) {
	query := `
		SELECT image_id, message_id, user_id, filename, original_filename,
			   file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
		FROM chat_images
		WHERE user_id = ? AND uploaded_at BETWEEN ? AND ?
		ORDER BY uploaded_at DESC
	`

	rows, err := r.DB.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query images by date range: %w", err)
	}
	defer rows.Close()

	var images []*models.ChatImage
	for rows.Next() {
		img := &models.ChatImage{}
		err := rows.Scan(
			&img.ImageID,
			&img.MessageID,
			&img.UserID,
			&img.Filename,
			&img.OriginalFilename,
			&img.FilePath,
			&img.ThumbnailPath,
			&img.FileSize,
			&img.MimeType,
			&img.Width,
			&img.Height,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating images by date: %w", err)
	}

	return images, nil
}

// CanAccessImage checks if a user can access a specific image
// User can access if they're either sender or receiver of the message associated with the image.
func (r *MessageRepository) CanAccessImage(imageID, userID string) (bool, error) {
	// Query joins chat_images to messages to check the sender/receiver IDs
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM chat_images ci
			JOIN messages m ON ci.message_id = m.message_id
			WHERE ci.image_id = ?
			  AND (m.sender_id = ? OR m.receiver_id = ?)
		)
	`
	var exists bool
	// Use QueryRow to execute the SELECT EXISTS query and scan the result into 'exists'
	err := r.DB.QueryRow(query, imageID, userID, userID).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			// This case is unlikely for EXISTS, but handled defensively
			return false, nil 
		}
		return false, fmt.Errorf("failed to verify image access: %w", err)
	}

	return exists, nil
}