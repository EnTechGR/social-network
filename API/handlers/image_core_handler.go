package handlers

import (
	"database/sql"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Image Core Service - Centralized image handling
// ============================================================================

type ImageCoreService struct {
	db          *sql.DB
	uploadsDir  string
	maxFileSize int64 // in bytes
}

type ImageMetadata struct {
	ImageID          string
	UploaderUserID   string
	Filename         string
	OriginalFilename string
	FilePath         string
	ThumbnailPath    string
	FileSize         int64
	MimeType         string
	Width            int
	Height           int
	UploadedAt       time.Time
}

func NewImageCoreService(db *sql.DB, uploadsDir string, maxFileSize int64) *ImageCoreService {
	return &ImageCoreService{
		db:          db,
		uploadsDir:  uploadsDir,
		maxFileSize: maxFileSize,
	}
}

// ============================================================================
// Core Image Processing Function (used by all contexts)
// ============================================================================

func (s *ImageCoreService) ProcessAndSaveImage(
	file multipart.File,
	header *multipart.FileHeader,
	userID string,
) (*ImageMetadata, error) {
	// Validate file size
	if header.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed: %d bytes", s.maxFileSize)
	}

	// Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	if !isAllowedMimeType(mimeType) {
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	// Decode image to get dimensions
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Reset file pointer for saving
	file.Seek(0, 0)

	// Generate unique filename
	imageID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = "." + format
	}
	filename := imageID + ext

	// Save original image
	filePath := filepath.Join(s.uploadsDir, filename)
	if err := s.saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Generate and save thumbnail
	thumbnailFilename := imageID + "_thumb" + ext
	thumbnailPath := filepath.Join(s.uploadsDir, thumbnailFilename)
	if err := s.generateThumbnail(filePath, thumbnailPath, 200, 200); err != nil {
		// Clean up original file if thumbnail fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Create metadata object
	metadata := &ImageMetadata{
		ImageID:          imageID,
		UploaderUserID:   userID,
		Filename:         filename,
		OriginalFilename: header.Filename,
		FilePath:         filePath,
		ThumbnailPath:    thumbnailPath,
		FileSize:         header.Size,
		MimeType:         mimeType,
		Width:            width,
		Height:           height,
		UploadedAt:       time.Now(),
	}

	return metadata, nil
}

// ============================================================================
// Insert image metadata into images_core table
// ============================================================================

func (s *ImageCoreService) SaveImageMetadata(tx *sql.Tx, metadata *ImageMetadata) error {
	query := `
		INSERT INTO images_core (
			image_id, uploader_user_id, filename, original_filename,
			file_path, thumbnail_path, file_size, mime_type, 
			width, height, uploaded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(query,
		metadata.ImageID,
		metadata.UploaderUserID,
		metadata.Filename,
		metadata.OriginalFilename,
		metadata.FilePath,
		metadata.ThumbnailPath,
		metadata.FileSize,
		metadata.MimeType,
		metadata.Width,
		metadata.Height,
		metadata.UploadedAt,
	)

	return err
}

// ============================================================================
// Context-Specific Image Handlers
// ============================================================================

// 1. USER AVATAR UPLOAD
func (s *ImageCoreService) UploadUserAvatar(
	file multipart.File,
	header *multipart.FileHeader,
	userID string,
) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing avatar if any (for cleanup)
	var oldImageID sql.NullString
	tx.QueryRow(`SELECT image_id FROM user_avatars WHERE user_id = ?`, userID).Scan(&oldImageID)

	// Process and save image
	metadata, err := s.ProcessAndSaveImage(file, header, userID)
	if err != nil {
		return err
	}

	// Save to images_core
	if err := s.SaveImageMetadata(tx, metadata); err != nil {
		// Clean up files on DB error
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to save image metadata: %w", err)
	}

	// Link to user (UPSERT pattern)
	query := `
		INSERT INTO user_avatars (user_id, image_id, set_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			image_id = excluded.image_id,
			set_at = excluded.set_at
	`
	if _, err := tx.Exec(query, userID, metadata.ImageID, time.Now()); err != nil {
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to link avatar to user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Soft delete old avatar (async cleanup job will handle file deletion)
	if oldImageID.Valid {
		s.db.Exec(`
			UPDATE images_core 
			SET deleted_at = CURRENT_TIMESTAMP 
			WHERE image_id = ?
		`, oldImageID.String)
	}

	return nil
}

// 2. POST IMAGE UPLOAD (supports multiple images)
func (s *ImageCoreService) UploadPostImages(
	files []*multipart.FileHeader,
	postID string,
	userID string,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current max display order for this post
	var maxOrder int
	tx.QueryRow(`
		SELECT COALESCE(MAX(display_order), 0) 
		FROM post_images 
		WHERE post_id = ?
	`, postID).Scan(&maxOrder)

	var savedImages []string // for cleanup on error
	displayOrder := maxOrder + 1

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			// Cleanup previously saved files
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Process image
		metadata, err := s.ProcessAndSaveImage(file, fileHeader, userID)
		if err != nil {
			s.cleanupFiles(savedImages)
			return err
		}
		savedImages = append(savedImages, metadata.FilePath, metadata.ThumbnailPath)

		// Save to images_core
		if err := s.SaveImageMetadata(tx, metadata); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to save image metadata: %w", err)
		}

		// Link to post
		postImageID := uuid.New().String()
		query := `
			INSERT INTO post_images (post_image_id, post_id, image_id, display_order, created_at)
			VALUES (?, ?, ?, ?, ?)
		`
		if _, err := tx.Exec(query, postImageID, postID, metadata.ImageID, displayOrder, time.Now()); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to link image to post: %w", err)
		}

		displayOrder++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.cleanupFiles(savedImages)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// 3. COMMENT IMAGE UPLOAD (single image)
func (s *ImageCoreService) UploadCommentImage(
	file multipart.File,
	header *multipart.FileHeader,
	commentID string,
	userID string,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process image
	metadata, err := s.ProcessAndSaveImage(file, header, userID)
	if err != nil {
		return err
	}

	// Save to images_core
	if err := s.SaveImageMetadata(tx, metadata); err != nil {
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to save image metadata: %w", err)
	}

	// Link to comment
	commentImageID := uuid.New().String()
	query := `
		INSERT INTO comment_images (comment_image_id, comment_id, image_id, created_at)
		VALUES (?, ?, ?, ?)
	`
	if _, err := tx.Exec(query, commentImageID, commentID, metadata.ImageID, time.Now()); err != nil {
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to link image to comment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		os.Remove(metadata.FilePath)
		os.Remove(metadata.ThumbnailPath)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// 4. CHAT MESSAGE IMAGE UPLOAD (supports multiple images)
func (s *ImageCoreService) UploadMessageImages(
	files []*multipart.FileHeader,
	messageID string,
	userID string,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var savedImages []string
	displayOrder := 1

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		metadata, err := s.ProcessAndSaveImage(file, fileHeader, userID)
		if err != nil {
			s.cleanupFiles(savedImages)
			return err
		}
		savedImages = append(savedImages, metadata.FilePath, metadata.ThumbnailPath)

		if err := s.SaveImageMetadata(tx, metadata); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to save image metadata: %w", err)
		}

		messageImageID := uuid.New().String()
		query := `
			INSERT INTO message_images (message_image_id, message_id, image_id, display_order, created_at)
			VALUES (?, ?, ?, ?, ?)
		`
		if _, err := tx.Exec(query, messageImageID, messageID, metadata.ImageID, displayOrder, time.Now()); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to link image to message: %w", err)
		}

		displayOrder++
	}

	if err := tx.Commit(); err != nil {
		s.cleanupFiles(savedImages)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// 5. GROUP MESSAGE IMAGE UPLOAD (supports multiple images)
func (s *ImageCoreService) UploadGroupMessageImages(
	files []*multipart.FileHeader,
	groupMessageID string,
	userID string,
) error {
	// Very similar to UploadMessageImages, just different relationship table
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var savedImages []string
	displayOrder := 1

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		metadata, err := s.ProcessAndSaveImage(file, fileHeader, userID)
		if err != nil {
			s.cleanupFiles(savedImages)
			return err
		}
		savedImages = append(savedImages, metadata.FilePath, metadata.ThumbnailPath)

		if err := s.SaveImageMetadata(tx, metadata); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to save image metadata: %w", err)
		}

		groupMessageImageID := uuid.New().String()
		query := `
			INSERT INTO group_message_images (group_message_image_id, group_message_id, image_id, display_order, created_at)
			VALUES (?, ?, ?, ?, ?)
		`
		if _, err := tx.Exec(query, groupMessageImageID, groupMessageID, metadata.ImageID, displayOrder, time.Now()); err != nil {
			s.cleanupFiles(savedImages)
			return fmt.Errorf("failed to link image to group message: %w", err)
		}

		displayOrder++
	}

	if err := tx.Commit(); err != nil {
		s.cleanupFiles(savedImages)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ============================================================================
// Image Retrieval Functions
// ============================================================================

type ImageInfo struct {
	ImageID       string
	FilePath      string
	ThumbnailPath string
	MimeType      string
	Width         int
	Height        int
	DisplayOrder  int // only for multi-image contexts
}

// Get user avatar
func (s *ImageCoreService) GetUserAvatar(userID string) (*ImageInfo, error) {
	query := `
		SELECT ic.image_id, ic.file_path, ic.thumbnail_path, ic.mime_type, ic.width, ic.height
		FROM user_avatars ua
		JOIN images_core ic ON ua.image_id = ic.image_id
		WHERE ua.user_id = ? AND ic.deleted_at IS NULL
	`

	var img ImageInfo
	err := s.db.QueryRow(query, userID).Scan(
		&img.ImageID, &img.FilePath, &img.ThumbnailPath,
		&img.MimeType, &img.Width, &img.Height,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No avatar
	}
	if err != nil {
		return nil, err
	}

	return &img, nil
}

// Get post images
func (s *ImageCoreService) GetPostImages(postID string) ([]ImageInfo, error) {
	query := `
		SELECT ic.image_id, ic.file_path, ic.thumbnail_path, ic.mime_type, 
		       ic.width, ic.height, pi.display_order
		FROM post_images pi
		JOIN images_core ic ON pi.image_id = ic.image_id
		WHERE pi.post_id = ? AND ic.deleted_at IS NULL
		ORDER BY pi.display_order
	`

	rows, err := s.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []ImageInfo
	for rows.Next() {
		var img ImageInfo
		err := rows.Scan(
			&img.ImageID, &img.FilePath, &img.ThumbnailPath,
			&img.MimeType, &img.Width, &img.Height, &img.DisplayOrder,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

// Get comment image
func (s *ImageCoreService) GetCommentImage(commentID string) (*ImageInfo, error) {
	query := `
		SELECT ic.image_id, ic.file_path, ic.thumbnail_path, ic.mime_type, ic.width, ic.height
		FROM comment_images ci
		JOIN images_core ic ON ci.image_id = ic.image_id
		WHERE ci.comment_id = ? AND ic.deleted_at IS NULL
	`

	var img ImageInfo
	err := s.db.QueryRow(query, commentID).Scan(
		&img.ImageID, &img.FilePath, &img.ThumbnailPath,
		&img.MimeType, &img.Width, &img.Height,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &img, nil
}

// ============================================================================
// Image Deletion Functions
// ============================================================================

// Delete user avatar (soft delete)
func (s *ImageCoreService) DeleteUserAvatar(userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get image_id
	var imageID string
	err = tx.QueryRow(`SELECT image_id FROM user_avatars WHERE user_id = ?`, userID).Scan(&imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No avatar to delete
		}
		return err
	}

	// Delete relationship
	if _, err := tx.Exec(`DELETE FROM user_avatars WHERE user_id = ?`, userID); err != nil {
		return err
	}

	// Soft delete image
	if _, err := tx.Exec(`
		UPDATE images_core 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE image_id = ?
	`, imageID); err != nil {
		return err
	}

	return tx.Commit()
}

// ============================================================================
// Background Cleanup Job
// ============================================================================

// Clean up soft-deleted images (should be run periodically)
func (s *ImageCoreService) CleanupDeletedImages(olderThanDays int) error {
	// Find images deleted more than X days ago
	query := `
		SELECT image_id, file_path, thumbnail_path
		FROM images_core
		WHERE deleted_at IS NOT NULL
		AND deleted_at < datetime('now', '-' || ? || ' days')
	`

	rows, err := s.db.Query(query, olderThanDays)
	if err != nil {
		return err
	}
	defer rows.Close()

	var imagesToDelete []struct {
		ImageID       string
		FilePath      string
		ThumbnailPath string
	}

	for rows.Next() {
		var img struct {
			ImageID       string
			FilePath      string
			ThumbnailPath string
		}
		if err := rows.Scan(&img.ImageID, &img.FilePath, &img.ThumbnailPath); err != nil {
			continue
		}
		imagesToDelete = append(imagesToDelete, img)
	}

	// Delete files and database records
	for _, img := range imagesToDelete {
		// Delete files from filesystem
		os.Remove(img.FilePath)
		os.Remove(img.ThumbnailPath)

		// Delete from database
		s.db.Exec(`DELETE FROM images_core WHERE image_id = ?`, img.ImageID)
	}

	return nil
}

// ============================================================================
// Utility Functions
// ============================================================================

func (s *ImageCoreService) saveFile(src io.Reader, dst string) error {
	// Ensure directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, src)
	return err
}

func (s *ImageCoreService) generateThumbnail(srcPath, dstPath string, maxWidth, maxHeight uint) error {
	// Open source image
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode image and get format
	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	newWidth, newHeight := calculateThumbnailSize(origWidth, origHeight, int(maxWidth), int(maxHeight))

	// Create thumbnail using nearest neighbor (fast, standard library only)
	thumbnail := resizeImage(img, newWidth, newHeight)

	// Create output file
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode based on original format
	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(out, thumbnail)
	case "gif":
		return gif.Encode(out, thumbnail, nil)
	default:
		// Default to JPEG for unknown formats
		return jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: 85})
	}
}

func (s *ImageCoreService) cleanupFiles(filePaths []string) {
	for _, path := range filePaths {
		os.Remove(path)
	}
}

func isAllowedMimeType(mimeType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		// Note: WebP requires external package: golang.org/x/image/webp
		// Uncomment the line below and add the import if you need WebP support
		// "image/webp",
	}

	for _, t := range allowed {
		if t == mimeType {
			return true
		}
	}
	return false
}

// calculateThumbnailSize calculates new dimensions maintaining aspect ratio
func calculateThumbnailSize(origWidth, origHeight, maxWidth, maxHeight int) (int, int) {
	if origWidth <= maxWidth && origHeight <= maxHeight {
		return origWidth, origHeight
	}

	widthRatio := float64(maxWidth) / float64(origWidth)
	heightRatio := float64(maxHeight) / float64(origHeight)

	// Use the smaller ratio to ensure image fits within bounds
	ratio := widthRatio
	if heightRatio < widthRatio {
		ratio = heightRatio
	}

	newWidth := int(float64(origWidth) * ratio)
	newHeight := int(float64(origHeight) * ratio)

	return newWidth, newHeight
}

// resizeImage performs nearest-neighbor image resizing (standard library only)
func resizeImage(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Nearest neighbor resampling
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map destination pixel to source pixel
			srcX := (x * srcWidth) / width
			srcY := (y * srcHeight) / height

			// Copy pixel
			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}