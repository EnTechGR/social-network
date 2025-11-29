package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forum/middleware"
	"forum/models"
	"forum/repository/message"
	"forum/utils"
	"forum/websocket"
	// You will need a package for image processing/metadata
)

const maxUploadSize = 5 * 1024 * 1024 // 5MB limit
const uploadPath = "./uploads/chat_images"

// ChatImageHandler handles file upload and serving for chat images
type ChatImageHandler struct {
	MessageRepo *message.MessageRepository
	Hub         *websocket.Hub
}

// NewChatImageHandler creates a new ChatImageHandler
func NewChatImageHandler(messageRepo *message.MessageRepository, hub *websocket.Hub) *ChatImageHandler {
	// Ensure the upload directory exists
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	return &ChatImageHandler{
		MessageRepo: messageRepo,
		Hub:         hub,
	}
}

// UploadChatImage handles the multipart form upload for a new message with an image
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
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		utils.ErrorResponse(w, fmt.Sprintf("File too large or invalid form: %v", err), http.StatusBadRequest)
		return
	}

	// 2. Extract Fields (Receiver ID and Content/Caption)
	receiverID := strings.TrimSpace(r.FormValue("receiver_id"))
	caption := strings.TrimSpace(r.FormValue("content"))

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
	defer file.Close()
	

	// 4. Generate metadata and file path
	imageID := utils.GenerateUUID()
	filename := imageID + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadPath, filename)

	// NOTE: Using dummy values for image processing (replace with real logic if available)
	const DUMMY_WIDTH = 800
	const DUMMY_HEIGHT = 600
	thumbnailPath := filepath.Join(uploadPath, "thumb_"+filename) // Dummy thumbnail path

	// 5. Save the file to disk (copy from form file to new file)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file on disk: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

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
		FilePath:         strings.TrimPrefix(filePath, "./"),
		ThumbnailPath:    strings.TrimPrefix(thumbnailPath, "./"),
		FileSize:         header.Size,
		MimeType:         header.Header.Get("Content-Type"),
		Width:            DUMMY_WIDTH,
		Height:           DUMMY_HEIGHT,
		UploadedAt:       now,
	}

	// 7. Save to Database (Transactional)
	if err := h.MessageRepo.CreateMessageAndImage(message, image); err != nil {
		log.Printf("Failed to create message and image transactionally: %v", err)
		// Clean up the file on disk since DB save failed
		os.Remove(filePath)
		os.Remove(thumbnailPath)
		utils.ErrorResponse(w, "Failed to save message and image record", http.StatusInternalServerError)
		return
	}

	// 8. Prepare final message model for response and WS broadcast
	messageWithImage := models.MessageWithUser{
		MessageID:  message.MessageID,
		SenderID:   message.SenderID,
		SenderName: user.Username,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
		IsRead:     message.IsRead,
		// ✅ CRUCIAL: Embed the image data
		Image: image,
	}

	// ✅ DEBUG: Log the message before sending
	log.Printf("[DEBUG] Image message created: MessageID=%s, ImageID=%s, FilePath=%s",
		messageWithImage.MessageID,
		messageWithImage.Image.ImageID,
		messageWithImage.Image.FilePath)

	// 9. Broadcast message via WebSocket to the receiver
	if h.Hub != nil {
		// Use the unified chat message function
		h.Hub.SendChatMessage(receiverID, messageWithImage)
	}

	// 10. Send HTTP success response (for sender's immediate display)
	utils.JSONResponse(w, map[string]interface{}{
		"message": messageWithImage, // Return the message with the image data
	}, http.StatusCreated)
}

// ServeChatImage handles serving the uploaded image file (requires access check)
// ServeChatImage handles serving the image file, but only after an access check.
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

// GetMessageImages will be covered in the next step (when you implement the message retrieval update)
// GetConversationGallery will be covered later

// DeleteChatImage deletes an image and its associated message
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

	// Get image ID from URL path (e.g., /forum/api/chat/images/delete/image-id-here)
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
		utils.ErrorResponse(w, "Image not found", http.StatusNotFound)
		return
	}

	// 2. Check if the current user is the sender/owner of the message
	msg, err := h.MessageRepo.GetByID(img.MessageID)
	if err != nil {
		log.Printf("Message not found for image %s: %v", imageID, err)
		utils.ErrorResponse(w, "Associated message not found", http.StatusNotFound)
		return
	}

	if msg.SenderID != user.ID {
		utils.ErrorResponse(w, "Unauthorized: Only the sender can delete the image and message", http.StatusForbidden)
		return
	}

	// 3. Delete from filesystem
	if err := os.Remove(img.FilePath); err != nil {
		// Log error but continue to try and clean up database records
		log.Printf("Failed to delete file from disk %s: %v", img.FilePath, err)
	}

	// 4. Delete image record from database
	if err := h.MessageRepo.DeleteChatImage(imageID); err != nil {
		log.Printf("Failed to delete chat image record: %v", err)
		utils.ErrorResponse(w, "Failed to delete image record", http.StatusInternalServerError)
		return
	}

	// 5. Delete associated message record
	if err := h.MessageRepo.Delete(img.MessageID, user.ID); err != nil {
		// NOTE: This call relies on MessageRepo.Delete checking if user.ID is the sender
		log.Printf("Failed to delete message record: %v", err)
		utils.ErrorResponse(w, "Failed to delete associated message", http.StatusInternalServerError)
		return
	}

	// 6. Notify receiver via WebSocket that message was deleted
	if h.Hub != nil {
		h.Hub.SendMessageDeleteNotification(msg.ReceiverID, img.MessageID, user.ID)
	}

	utils.JSONResponse(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

// GetUserImageStats retrieves total image count and size for a user
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

	count, err := h.MessageRepo.GetImageCountByUser(user.ID)
	if err != nil {
		log.Printf("Failed to get image count: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve image count", http.StatusInternalServerError)
		return
	}

	totalSize, err := h.MessageRepo.GetTotalImagesSizeByUser(user.ID)
	if err != nil {
		log.Printf("Failed to get total image size: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve image size", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"image_count":      count,
		"total_size_bytes": totalSize,
	}, http.StatusOK)
}

func (h *ChatImageHandler) GetMessageImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: Implementation skipped - this is a placeholder to fix compilation.
	// Logic would typically involve validating the user, getting conversation ID or other user ID
	// from query params, and fetching image metadata for those messages.

	log.Println("Note: GetMessageImages handler called but not fully implemented.")
	utils.JSONResponse(w, map[string]interface{}{
		"images":  []interface{}{}, // Return empty list for now
		"success": true,
	}, http.StatusOK)
}

// GetConversationGallery retrieves all images from a conversation between two users
func (h *ChatImageHandler) GetConversationGallery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: Implementation skipped - this is a placeholder to fix compilation.
	// Logic would involve getting the other user's ID from query params,
	// validating the current user, and then calling MessageRepo.GetConversationImages(userID, otherUserID).

	log.Println("Note: GetConversationGallery handler called but not fully implemented.")
	utils.JSONResponse(w, map[string]interface{}{
		"gallery": []interface{}{}, // Return empty list for now
		"success": true,
	}, http.StatusOK)
}
