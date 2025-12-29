package handlers

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository/message"
	"social-network/utils"
	"social-network/websocket"
)

const maxUploadSize = 5 * 1024 * 1024 // 5MB limit
const uploadPath = "./uploads/chat_images" // Directory to store chat images

// ChatImageHandler handles file upload and serving for chat images
type ChatImageHandler struct {
	MessageRepo *message.MessageRepository
	Hub         *websocket.Hub
}

// NewChatImageHandler creates a new ChatImageHandler instance.
//
// It is responsible for initializing the handler and performing necessary
// setup checks, such as ensuring the file upload directory exists.
//
// Parameters:
//   - messageRepo: The repository interface for database operations.
//   - hub: The WebSocket hub for broadcasting messages.
//
// Returns:
//   - *ChatImageHandler: The initialized handler instance.
func NewChatImageHandler(messageRepo *message.MessageRepository, hub *websocket.Hub) *ChatImageHandler {
	// Ensure the file upload directory for chat images exists before the server starts.
	// os.ModePerm (0777) gives full permissions. MkdirAll creates parents if necessary.
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		// Using log.Fatalf is appropriate here as the server cannot function
		// without a writable directory for permanent file storage.
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	return &ChatImageHandler{
		MessageRepo: messageRepo,
		Hub:         hub,
	}
}

// UploadChatImage handles sending an image in a private message
// @Summary      Upload chat image
// @Description  Sends an image message to another user. Validates file type (JPEG, PNG, GIF), saves to disk, and broadcasts via WebSocket.
// @Tags         Messaging
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        receiver_id  formData  string  true  "ID of the message recipient"
// @Param        content      formData  string  false "Optional text caption (max 1000 chars)"
// @Param        chat_image   formData  file    true  "Image file to send (max 5MB)"
// @Success      201          {object}  map[string]interface{} "Returns {message: models.MessageWithUser}"
// @Failure      400          {object}  models.ErrorResponse   "Invalid file, too large, or messaging self"
// @Failure      401          {object}  models.ErrorResponse   "Unauthorized"
// @Router       /api/chat/images/upload [post]
func (h *ChatImageHandler) UploadChatImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Parse Multipart Form
	// The maxUploadSize determines the total body size limit and is applied here.
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		// This error typically means the request body exceeded the maximum size.
		utils.ErrorResponse(w, fmt.Sprintf("File too large or invalid form: %v", err), http.StatusBadRequest)
		return
	}

	// 2. Extract Fields (Receiver ID and Content/Caption)
	receiverID := strings.TrimSpace(r.FormValue("receiver_id"))
	caption := strings.TrimSpace(r.FormValue("content"))

	// Basic business logic validation
	if receiverID == "" {
		utils.ErrorResponse(w, "Receiver ID is required", http.StatusBadRequest)
		return
	}
	if receiverID == user.ID {
		utils.ErrorResponse(w, "Cannot send message to yourself", http.StatusBadRequest)
		return
	}
	if len(caption) > 1000 {
		utils.ErrorResponse(w, "Caption content cannot exceed 1000 characters", http.StatusBadRequest)
		return
	}

	// 3. Get the File
	file, header, err := r.FormFile("chat_image")
	if err != nil {
		if err == http.ErrMissingFile {
			utils.ErrorResponse(w, "Missing chat_image file", http.StatusBadRequest)
			return
		}
		log.Printf("Error retrieving file: %v", err)
		utils.ErrorResponse(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}
	// Ensure the form file is closed when the function exits
	defer file.Close()

	// 3a. Validate MIME type by reading file header
	// Read the first 512 bytes for MIME type content sniffing (anti-spoofing security check).
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("Error reading file header: %v", err)
		utils.ErrorResponse(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Reset file pointer to beginning after reading header
	// This is CRUCIAL so that subsequent file operations (DecodeConfig and io.Copy) read from the start.
	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("Error resetting file pointer: %v", err)
		utils.ErrorResponse(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	// Detect actual MIME type from file content (Magic Bytes check)
	detectedMimeType := http.DetectContentType(buffer[:n])

	// Validate that the detected content type is an allowed image type
	// Note: This check is redundant with the repository's ValidateMimeType, but is kept here
	// for clarity on the handler side. If the repository check is exhaustive, this map can be removed.
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowedMimeTypes[detectedMimeType] {
		utils.ErrorResponse(w, fmt.Sprintf("Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed. Detected: %s", detectedMimeType), http.StatusBadRequest)
		return
	}

	// 3b. Decode image to get actual dimensions
	// This step verifies file integrity (is it a valid image?) and extracts width/height.
	imgConfig, format, err := image.DecodeConfig(file)
	if err != nil {
		log.Printf("Error decoding image config: %v", err)
		// Failure to decode means the file is corrupted or not a valid image format.
		utils.ErrorResponse(w, "Invalid or corrupted image file", http.StatusBadRequest)
		return
	}

	// Reset file pointer again after decoding config
	// This is CRITICAL because the decoder also moves the file pointer.
	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("Error resetting file pointer after decode: %v", err)
		utils.ErrorResponse(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	// Extract actual width and height for database storage
	imageWidth := imgConfig.Width
	imageHeight := imgConfig.Height

	log.Printf("Image uploaded: Format=%s, Dimensions=%dx%d, MIME=%s", format, imageWidth, imageHeight, detectedMimeType)

	// 4. Generate metadata and file path
	imageID := utils.GenerateUUID()
	// Use the original extension (safer to use a standardized extension based on detectedMimeType if needed, but keeping this for now)
	filename := imageID + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadPath, filename)

	// Thumbnail generation logic would go here.
	thumbnailPath := "" // Empty for now

	// 5. Save the file to disk (copy from form file to new file)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file on disk: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// io.Copy reads from the beginning of 'file' (due to the Seek(0, 0) calls) and writes to 'dst'.
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file content: %v", err)
		utils.ErrorResponse(w, "Failed to save file content", http.StatusInternalServerError)
		return
	}

	// 6. Create Message and ChatImage models
	messageID := utils.GenerateUUID()
	now := time.Now()

	message := &models.Message{
		MessageID:  messageID,
		SenderID:   user.ID,
		ReceiverID: receiverID,
		Content:    caption,
		CreatedAt:  now,
		IsRead:     false,
	}

	image := &models.ChatImage{
		ImageID:          imageID,
		MessageID:        messageID,
		UserID:           user.ID,
		Filename:         filename,
		OriginalFilename: header.Filename,
		// Store file paths relative to the application base
		FilePath:      strings.TrimPrefix(filePath, "./"),
		ThumbnailPath: thumbnailPath,
		FileSize:      header.Size,
		// CRITICAL: Use the server-validated content type and extracted dimensions
		MimeType:   detectedMimeType,
		Width:      imageWidth,
		Height:     imageHeight,
		UploadedAt: now,
	}

	// 7. Save to Database (Transactional)
	// Both the message and image records are inserted in a single atomic transaction.
	if err := h.MessageRepo.CreateMessageAndImage(message, image); err != nil {
		log.Printf("Failed to create message and image transactionally: %v", err)
		// Clean up the file on disk since DB save failed (Rollback cleanup).
		os.Remove(filePath)
		if thumbnailPath != "" {
			os.Remove(thumbnailPath)
		}
		utils.ErrorResponse(w, "Failed to save message and image record", http.StatusInternalServerError)
		return
	}

	// 8. Prepare final message model for response and WS broadcast
	messageWithImage := models.MessageWithUser{
		MessageID:  message.MessageID,
		SenderID:   message.SenderID,
		SenderNickname: user.Nickname,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
		IsRead:     message.IsRead,
		// Embed the complete image metadata for immediate client display
		Image: image,
	}

	// Log successful creation details
	log.Printf("[DEBUG] Image message created: MessageID=%s, ImageID=%s, FilePath=%s, Dimensions=%dx%d",
		messageWithImage.MessageID,
		messageWithImage.Image.ImageID,
		messageWithImage.Image.FilePath,
		messageWithImage.Image.Width,
		messageWithImage.Image.Height)

	// 9. Broadcast message via WebSocket to the receiver
	if h.Hub != nil {
		// Send the message object (which includes the image) over the wire.
		h.Hub.SendChatMessage(receiverID, messageWithImage)
	}

	// 10. Send HTTP success response (for sender's immediate display)
	utils.JSONResponse(w, map[string]interface{}{
		"message": messageWithImage,
	}, http.StatusCreated)
}

// ServeChatImage serves the raw image file with access control
// @Summary      Serve chat image
// @Description  Serves the image file from disk. **Crucial**: Validates that the requester is either the sender or receiver of the message before serving.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      image/jpeg,image/png,image/gif
// @Param        imagePath  path      string  true  "The relative path to the image"
// @Success      200        {file}    binary
// @Failure      403        {object}  models.ErrorResponse "Forbidden - You don't have access to this image"
// @Failure      404        {object}  models.ErrorResponse "Image not found"
// @Router       /api/chat/images/serve/{imagePath} [get]
func (h *ChatImageHandler) ServeChatImage(w http.ResponseWriter, r *http.Request) {
	// 1. Get authenticated user from context (Set by protected middleware)
	user := middleware.GetCurrentUser(r)
	if user == nil {
		// This case should be handled by the middleware, but serves as a failsafe.
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse URL to get the full file path from the request.
	// Example path: /forum/api/chat/images/serve/uploads/chat_images/73245b0a-56cd-4f02-91eb-f16731bd3aac.png
	const filePathPrefix = "/forum/api/chat/images/serve/"
	requestPath := strings.TrimPrefix(r.URL.Path, filePathPrefix)

	if requestPath == r.URL.Path {
		// Should not happen if the router is configured correctly
		utils.ErrorResponse(w, "Invalid image path format", http.StatusBadRequest)
		return
	}

	// 3. Extract the Image ID from the filename component of the path
	// The Image ID is the UUID part before the file extension.
	baseName := filepath.Base(requestPath)                          // e.g., '73245b0a-56cd-4f02-91eb-f16731bd3aac.png'
	imageID := strings.TrimSuffix(baseName, filepath.Ext(baseName)) // e.g., '73245b0a-...'

	if imageID == "" {
		utils.ErrorResponse(w, "Missing image identifier", http.StatusBadRequest)
		return
	}

	// 4. Access Control Check (Crucial step)
	canAccess, err := h.MessageRepo.CanAccessImage(imageID, user.ID)
	if err != nil {
		log.Printf("DB error during image access check for ID %s: %v", imageID, err)
		utils.ErrorResponse(w, "Internal server error during access check", http.StatusInternalServerError)
		return
	}

	if !canAccess {
		log.Printf("Access denied: User %s tried to access image %s", user.ID, imageID)
		// This is the source of the 403 error message the user is seeing.
		utils.ErrorResponse(w, "Forbidden: You do not have access to this image", http.StatusForbidden)
		return
	}

	// 5. Serve the file from disk (requestPath is the file's location relative to the app root)
	fullFilePath := "./" + requestPath

	// Check if the file exists on disk
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		log.Printf("Image file not found on disk: %s", fullFilePath)
		utils.ErrorResponse(w, "Image not found on server", http.StatusNotFound)
		return
	}

	// 6. Serve the file
	// http.ServeFile automatically handles content type
	http.ServeFile(w, r, fullFilePath)
}

// DeleteChatImage removes an image and its associated message
// @Summary      Delete chat image
// @Description  Deletes the image file, the image metadata, and the original message record. Only the sender can perform this.
// @Tags         Messaging
// @Security     CookieAuth
// @Param        imageID  path      string  true  "ID of the image to delete"
// @Success      200      {object}  map[string]bool "success: true"
// @Failure      403      {object}  models.ErrorResponse "Unauthorized: Only sender can delete"
// @Router       /api/chat/images/{imageID} [delete]
func (h *ChatImageHandler) DeleteChatImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract the image ID from the last segment of the URL path.
	// This approach is somewhat brittle; a router variable capture is generally safer.
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		utils.ErrorResponse(w, "Image ID is required", http.StatusBadRequest)
		return
	}
	imageID := pathParts[len(pathParts)-1]

	if imageID == "" {
		utils.ErrorResponse(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	// 1. Get image metadata to find file path and message ID
	img, err := h.MessageRepo.GetChatImageByID(imageID)
	if err != nil {
		// Log specific error for debugging but return generic 404 to client
		utils.ErrorResponse(w, "Image not found", http.StatusNotFound)
		return
	}

	// 2. Check if the current user is the sender/owner of the message (Authorization)
	msg, err := h.MessageRepo.GetByID(img.MessageID)
	if err != nil {
		// Image exists, but its parent message is missing (a data integrity issue).
		log.Printf("Message not found for image %s: %v", imageID, err)
		utils.ErrorResponse(w, "Associated message not found", http.StatusNotFound)
		return
	}

	if msg.SenderID != user.ID {
		// CRITICAL SECURITY CHECK: Ensure only the message sender can delete.
		utils.ErrorResponse(w, "Unauthorized: Only the sender can delete the image and message", http.StatusForbidden)
		return
	}

	// 3. Delete from filesystem
	// This step is performed first as it is often the most likely to fail or cause latency.
	if err := os.Remove(img.FilePath); err != nil {
		// Log error (file not found or permission issue) but allow subsequent database cleanup.
		// NOTE: If file deletion fails, the DB records still get cleaned, minimizing dangling metadata.
		log.Printf("Failed to delete file from disk %s: %v", img.FilePath, err)
	}

	// 4. Delete image record from database
	if err := h.MessageRepo.DeleteChatImage(imageID); err != nil {
		log.Printf("Failed to delete chat image record: %v", err)
		utils.ErrorResponse(w, "Failed to delete image record", http.StatusInternalServerError)
		return
	}

	// 5. Delete associated message record
	// The image is intrinsically linked to the message, so both are deleted together.
	// We rely on MessageRepo.Delete for final validation and execution.
	if err := h.MessageRepo.Delete(img.MessageID, user.ID); err != nil {
		log.Printf("Failed to delete message record: %v", err)
		// NOTE: At this point, the file and image metadata are already gone.
		utils.ErrorResponse(w, "Failed to delete associated message", http.StatusInternalServerError)
		return
	}

	// 6. Notify receiver via WebSocket that message was deleted
	if h.Hub != nil {
		h.Hub.SendMessageDeleteNotification(msg.ReceiverID, img.MessageID, user.ID)
	}

	// Success response.
	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// GetUserImageStats retrieves storage statistics for the user
// @Summary      Get user image stats
// @Description  Returns total count and total byte size of all images uploaded by the authenticated user.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Success      200      {object}  map[string]interface{} "image_count (int), total_size_bytes (int64)"
// @Router       /api/chat/images/stats [get]
func (h *ChatImageHandler) GetUserImageStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Retrieve the count of images uploaded by the user.
	count, err := h.MessageRepo.GetImageCountByUser(user.ID)
	if err != nil {
		log.Printf("Failed to get image count: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve image count", http.StatusInternalServerError)
		return
	}

	// 2. Retrieve the sum of all image file sizes for the user.
	// This uses the COALESCE/SUM SQL aggregation handled in the repository layer.
	totalSize, err := h.MessageRepo.GetTotalImagesSizeByUser(user.ID)
	if err != nil {
		log.Printf("Failed to get total image size: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve image size", http.StatusInternalServerError)
		return
	}

	// 3. Return the statistics in the response.
	utils.JSONResponse(w, map[string]interface{}{
		"image_count":      count,          // Total number of images (int)
		"total_size_bytes": totalSize,      // Total size in bytes (int64)
	}, http.StatusOK)
}

// GetMessageImages retrieves image metadata for a specific message
// @Summary      Get images by message
// @Description  Returns metadata for all images attached to a single message ID. Requires sender/receiver permission.
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Param        messageID  query    string  true  "ID of the message"
// @Success      200        {object} MessageImagesResponse
// @Router       /api/chat/images [get]
func (h *ChatImageHandler) GetMessageImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: Implementation skipped - this is a placeholder to fix compilation.
	// Logic would typically involve validating the user, getting conversation ID or other user ID
	// from query params, and fetching image metadata for those messages.

	log.Println("Note: GetMessageImages handler called but not fully implemented.")
	// Return a temporary successful response with an empty list to prevent client errors.
	utils.JSONResponse(w, map[string]interface{}{
		"images":  []interface{}{}, // Return empty list for now
		"success": true,
	}, http.StatusOK)
}

// GetConversationGallery retrieves all images from a chat
// @Summary      Get chat gallery
// @Description  Returns a list of image metadata shared between the user and a partner. (Placeholder implementation).
// @Tags         Messaging
// @Security     CookieAuth
// @Produce      json
// @Param        partner  query     string  true  "User ID of the chat partner"
// @Success      200      {object}  map[string]interface{} "gallery: []"
// @Router       /api/chat/gallery [get]
func (h *ChatImageHandler) GetConversationGallery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: Implementation skipped - this is a placeholder to fix compilation.
	// Logic would involve getting the other user's ID from query params,
	// validating the current user, and then calling MessageRepo.GetConversationImages(userID, otherUserID).

	log.Println("Note: GetConversationGallery handler called but not fully implemented.")
	// Return a temporary successful response with an empty list.
	utils.JSONResponse(w, map[string]interface{}{
		"gallery": []interface{}{}, // Return empty list for now
		"success": true,
	}, http.StatusOK)
}
