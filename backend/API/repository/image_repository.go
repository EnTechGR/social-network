package repository

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
	"strings"
	"time"

	"social-network/models"
	"social-network/utils"

	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"
)

const (
	thumbnailSmallMaxWidth  uint = 320
	thumbnailSmallMaxHeight uint = 320
	thumbnailSmallJPEGQ          = 90

	thumbnailMediumMaxWidth  uint = 640
	thumbnailMediumMaxHeight uint = 640
	thumbnailMediumJPEGQ          = 92
)

// ============================================================================
// Enhanced Image Repository - Centralized image handling
// ============================================================================

type ImageRepository struct {
	db          *sql.DB
	uploadsDir  string
	maxFileSize int64 // in bytes
}

// ImageMetadata represents processed image information
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

// ImageInfo represents retrieved image information
type ImageInfo struct {
	ImageID       string
	FilePath      string
	ThumbnailPath string
	MimeType      string
	Width         int
	Height        int
	DisplayOrder  int // only for multi-image contexts
}

// NewImageRepository creates a new ImageRepository instance
func NewImageRepository(db *sql.DB) *ImageRepository {
	const defaultUploadsDir = "uploads"
	const defaultMaxFileSize = 20 << 20 // 20MB

	return &ImageRepository{
		db:          db,
		uploadsDir:  defaultUploadsDir,
		maxFileSize: defaultMaxFileSize,
	}
}

// ============================================================================
// OLD METHODS - Keep for backwards compatibility
// ============================================================================

// Create - OLD METHOD for backwards compatibility
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

// GetByPostID - OLD METHOD for backwards compatibility
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

// DeleteByPostID - OLD METHOD kept for backwards compatibility (uses old schema)
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

// ============================================================================
// NEW METHODS - Core Image Processing
// ============================================================================

// ProcessAndSaveImage processes an uploaded image file and saves it with thumbnail
func (r *ImageRepository) ProcessAndSaveImage(
	file multipart.File,
	header *multipart.FileHeader,
	userID string,
) (*ImageMetadata, error) {
	// Validate file size
	if header.Size > r.maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed: %d bytes", r.maxFileSize)
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
	filePath := filepath.Join(r.uploadsDir, filename)
	if err := r.saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Generate and save responsive thumbnails.
	thumbnailSmallFilename := imageID + "_thumb_sm" + ext
	thumbnailSmallPath := filepath.Join(r.uploadsDir, thumbnailSmallFilename)
	if err := r.generateThumbnail(filePath, thumbnailSmallPath, thumbnailSmallMaxWidth, thumbnailSmallMaxHeight, thumbnailSmallJPEGQ); err != nil {
		// Clean up original file if thumbnail fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to generate small thumbnail: %w", err)
	}

	thumbnailMediumFilename := imageID + "_thumb_md" + ext
	thumbnailMediumPath := filepath.Join(r.uploadsDir, thumbnailMediumFilename)
	if err := r.generateThumbnail(filePath, thumbnailMediumPath, thumbnailMediumMaxWidth, thumbnailMediumMaxHeight, thumbnailMediumJPEGQ); err != nil {
		// Clean up original file if thumbnail fails
		os.Remove(filePath)
		os.Remove(thumbnailSmallPath)
		return nil, fmt.Errorf("failed to generate medium thumbnail: %w", err)
	}

	// Create metadata object
	metadata := &ImageMetadata{
		ImageID:          imageID,
		UploaderUserID:   userID,
		Filename:         filename,
		OriginalFilename: header.Filename,
		FilePath:         filePath,
		ThumbnailPath:    thumbnailMediumPath,
		FileSize:         header.Size,
		MimeType:         mimeType,
		Width:            width,
		Height:           height,
		UploadedAt:       time.Now(),
	}

	return metadata, nil
}

// SaveImageMetadata inserts image metadata into images_core table
func (r *ImageRepository) SaveImageMetadata(tx *sql.Tx, metadata *ImageMetadata) error {
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
// CONTEXT-SPECIFIC UPLOAD METHODS
// ============================================================================

// UploadUserAvatar uploads and processes a user avatar image
func (r *ImageRepository) UploadUserAvatar(
	file multipart.File,
	header *multipart.FileHeader,
	userID string,
) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing avatar if any (for cleanup)
	var oldImageID sql.NullString
	tx.QueryRow(`SELECT image_id FROM user_avatars WHERE user_id = ?`, userID).Scan(&oldImageID)

	// Process and save image
	metadata, err := r.ProcessAndSaveImage(file, header, userID)
	if err != nil {
		return err
	}

	// Save to images_core
	if err := r.SaveImageMetadata(tx, metadata); err != nil {
		// Clean up files on DB error
		r.cleanupFiles(r.filePathsForImageMetadata(metadata))
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
		r.cleanupFiles(r.filePathsForImageMetadata(metadata))
		return fmt.Errorf("failed to link avatar to user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		r.cleanupFiles(r.filePathsForImageMetadata(metadata))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Soft delete old avatar (async cleanup job will handle file deletion)
	if oldImageID.Valid {
		r.db.Exec(`
			UPDATE images_core 
			SET deleted_at = CURRENT_TIMESTAMP 
			WHERE image_id = ?
		`, oldImageID.String)
	}

	return nil
}

// UploadPostImages uploads and processes multiple images for a post,
// saving metadata to images_core and linking each image to the post
// via the post_images relationship table.
//
// Replace the existing UploadPostImages function in
// backend/API/repository/image_repository.go with this implementation.
func (r *ImageRepository) UploadPostImages(
	files []*multipart.FileHeader,
	postID string,
	userID string,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var savedFiles []string // paths for cleanup on error

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Process image: validate, decode, save to disk, generate thumbnail
		metadata, err := r.ProcessAndSaveImage(file, fileHeader, userID)
		if err != nil {
			r.cleanupFiles(savedFiles)
			return err
		}
		savedFiles = append(savedFiles, r.filePathsForImageMetadata(metadata)...)

		// 1. Insert into centralized images_core table
		if err := r.SaveImageMetadata(tx, metadata); err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to save image metadata: %w", err)
		}

		// 2. Insert into post_images relationship table (this was the missing link)
		postImageID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO post_images (post_image_id, post_id, image_id, display_order, created_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, postImageID, postID, metadata.ImageID, i+1)
		if err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to link image %s to post %s: %w", metadata.ImageID, postID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.cleanupFiles(savedFiles)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UploadCommentImages uploads and processes multiple images for a comment,
// saving metadata to images_core and linking each image to the comment via
// the comment_images relationship table.
func (r *ImageRepository) UploadCommentImages(
	files []*multipart.FileHeader,
	commentID string,
	userID string,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var savedFiles []string // paths for cleanup on error

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		metadata, err := r.ProcessAndSaveImage(file, fileHeader, userID)
		if err != nil {
			r.cleanupFiles(savedFiles)
			return err
		}
		savedFiles = append(savedFiles, r.filePathsForImageMetadata(metadata)...)

		if err := r.SaveImageMetadata(tx, metadata); err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to save image metadata: %w", err)
		}

		commentImageID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO comment_images (comment_image_id, comment_id, image_id, display_order, created_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, commentImageID, commentID, metadata.ImageID, i+1)
		if err != nil {
			r.cleanupFiles(savedFiles)
			return fmt.Errorf("failed to link image %s to comment %s: %w", metadata.ImageID, commentID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.cleanupFiles(savedFiles)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ============================================================================
// IMAGE RETRIEVAL METHODS
// ============================================================================

// GetUserAvatar retrieves user's avatar information
func (r *ImageRepository) GetUserAvatar(userID string) (*ImageInfo, error) {
	query := `
		SELECT ic.image_id, ic.file_path, ic.thumbnail_path, ic.mime_type, ic.width, ic.height
		FROM user_avatars ua
		JOIN images_core ic ON ua.image_id = ic.image_id
		WHERE ua.user_id = ? AND ic.deleted_at IS NULL
	`

	var img ImageInfo
	err := r.db.QueryRow(query, userID).Scan(
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

// GetPostImages retrieves all images for a post
func (r *ImageRepository) GetPostImages(postID string) ([]ImageInfo, error) {
	query := `
		SELECT ic.image_id, ic.file_path, ic.thumbnail_path, ic.mime_type, 
		       ic.width, ic.height, pi.display_order
		FROM post_images pi
		JOIN images_core ic ON pi.image_id = ic.image_id
		WHERE pi.post_id = ? AND ic.deleted_at IS NULL
		ORDER BY pi.display_order
	`

	rows, err := r.db.Query(query, postID)
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

// ============================================================================
// IMAGE DELETION METHODS
// ============================================================================

// DeleteUserAvatar deletes a user's avatar (soft delete)
func (r *ImageRepository) DeleteUserAvatar(userID string) error {
	tx, err := r.db.Begin()
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

// DeletePostImages deletes all images for a post (soft delete)
func (r *ImageRepository) DeletePostImages(postID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all image IDs for this post
	rows, err := tx.Query(`SELECT image_id FROM post_images WHERE post_id = ?`, postID)
	if err != nil {
		return err
	}

	var imageIDs []string
	for rows.Next() {
		var imageID string
		if err := rows.Scan(&imageID); err != nil {
			rows.Close()
			return err
		}
		imageIDs = append(imageIDs, imageID)
	}
	rows.Close()

	// Delete all relationships
	if _, err := tx.Exec(`DELETE FROM post_images WHERE post_id = ?`, postID); err != nil {
		return err
	}

	// Soft delete all images
	for _, imageID := range imageIDs {
		if _, err := tx.Exec(`
			UPDATE images_core 
			SET deleted_at = CURRENT_TIMESTAMP 
			WHERE image_id = ?
		`, imageID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func (r *ImageRepository) saveFile(src io.Reader, dst string) error {
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

func (r *ImageRepository) generateThumbnail(srcPath, dstPath string, maxWidth, maxHeight uint, jpegQuality int) error {
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

	// Create thumbnail using a higher-quality scaler to avoid visibly blocky card images.
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
		return jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: jpegQuality})
	case "png":
		return png.Encode(out, thumbnail)
	case "gif":
		return gif.Encode(out, thumbnail, nil)
	default:
		// Default to JPEG for unknown formats
		return jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: jpegQuality})
	}
}

func (r *ImageRepository) filePathsForImageMetadata(metadata *ImageMetadata) []string {
	paths := []string{metadata.FilePath}
	paths = append(paths, thumbnailVariants(metadata.ThumbnailPath)...)
	return paths
}

func thumbnailVariants(thumbnailPath string) []string {
	if thumbnailPath == "" {
		return nil
	}

	variants := []string{thumbnailPath}

	if strings.Contains(thumbnailPath, "_thumb_md") {
		variants = append(variants, strings.Replace(thumbnailPath, "_thumb_md", "_thumb_sm", 1))
	}

	return variants
}

func (r *ImageRepository) cleanupFiles(filePaths []string) {
	for _, path := range filePaths {
		os.Remove(path)
	}
}

func isAllowedMimeType(mimeType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
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

// resizeImage performs high-quality image resizing.
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

	return dst
}

// ============================================================================
// CLEANUP JOB
// ============================================================================

// CleanupDeletedImages removes soft-deleted images older than specified days
func (r *ImageRepository) CleanupDeletedImages(olderThanDays int) error {
	// Find images deleted more than X days ago
	query := `
		SELECT image_id, file_path, thumbnail_path
		FROM images_core
		WHERE deleted_at IS NOT NULL
		AND deleted_at < datetime('now', '-' || ? || ' days')
	`

	rows, err := r.db.Query(query, olderThanDays)
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
		for _, thumbPath := range thumbnailVariants(img.ThumbnailPath) {
			os.Remove(thumbPath)
		}

		// Delete from database
		r.db.Exec(`DELETE FROM images_core WHERE image_id = ?`, img.ImageID)
	}

	return nil
}
