package models

import "time"

// --- Core Data Models ---

// Message represents a private message record as stored in the database.
// This structure is primarily used for CRUD operations against the 'messages' table.
type Message struct {
	// MessageID is the unique primary key for the message record (e.g., a UUID).
	MessageID string `json:"message_id"`
	// SenderID is the unique identifier of the user who sent the message.
	SenderID string `json:"sender_id"`
	// ReceiverID is the unique identifier of the user intended to receive the message.
	ReceiverID string `json:"receiver_id"`
	// Content holds the text body of the message.
	Content string `json:"content"`
	// CreatedAt records the timestamp when the message was persisted.
	CreatedAt time.Time `json:"created_at"`
	// IsRead indicates the read status from the receiver's perspective (true if read).
	IsRead bool `json:"is_read"`
	// Image is an optional field containing metadata for an attached image, if present.
	Image *ChatImage `json:"image,omitempty"`
}

// ChatImage represents the metadata for an image file attached to a chat message.
// This corresponds directly to the 'chat_images' database table.
type ChatImage struct {
	// ImageID is the unique primary key for the image metadata record.
	ImageID string `json:"image_id"`
	// MessageID is the Foreign Key linking this image to its parent message.
	MessageID string `json:"message_id"`
	// UserID is the ID of the user who uploaded the image (should match Message.SenderID).
	UserID string `json:"user_id"`
	// Filename is the unique, server-side generated name of the stored file.
	Filename string `json:"filename"`
	// OriginalFilename is the name the user originally used for the file.
	OriginalFilename string `json:"original_filename"`
	// FilePath is the full storage path to the original image file.
	FilePath string `json:"file_path"`
	// ThumbnailPath is the storage path to the optimized thumbnail version.
	ThumbnailPath string `json:"thumbnail_path"`
	// FileSize is the size of the original file in bytes.
	FileSize int64 `json:"file_size"`
	// MimeType specifies the file format (e.g., image/jpeg).
	MimeType string `json:"mime_type"`
	// Width is the image width in pixels.
	Width int `json:"width"`
	// Height is the image height in pixels.
	Height int `json:"height"`
	// UploadedAt records the timestamp when the image metadata was created.
	UploadedAt time.Time `json:"uploaded_at"`
}

// --- Request and Response Models ---

// MessageWithUser is the payload used for API responses, including the sender's/receiver's display information.
// This is typically the result of a database JOIN between 'messages' and 'users'.
type MessageWithUser struct {
	MessageID string `json:"message_id"`
	SenderID string `json:"sender_id"`
	// SenderName is the display name of the message sender.
	SenderNickname string `json:"sender_name"`
	ReceiverID string `json:"receiver_id"`
	// ReceiverName is the display name of the message receiver.
	ReceiverName string `json:"receiver_name"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsRead bool `json:"is_read"`
	// Image is an optional field containing attached image metadata.
	Image *ChatImage `json:"image,omitempty"`
}

// CreateMessageRequest defines the structure for an incoming API request to send a new message.
type CreateMessageRequest struct {
	// ReceiverID is the ID of the intended recipient (required field).
	ReceiverID string `json:"receiver_id" binding:"required"`
	// Content is the text body of the message (required field).
	Content string `json:"content" binding:"required"`
}

// Conversation represents a single chat conversation thread in a list view.
// It summarizes the interaction with one specific peer user.
type Conversation struct {
	// UserID is the ID of the peer user in the conversation (not the current user).
	UserID string `json:"user_id"`
	// Nickname is the display name of the peer user.
	Nickname string `json:"nickname"`
	// LastMessage is the content of the most recent message in the thread.
	LastMessage string `json:"last_message"`
	// LastMessageTime is the timestamp of the most recent message.
	LastMessageTime time.Time `json:"last_message_time"`
	// UnreadCount is the total number of unread messages in this thread for the current user.
	UnreadCount int `json:"unread_count"`
	// IsOnline indicates the current online status of the peer user.
	IsOnline bool `json:"is_online"`
}

// MessageResponse defines the standard API response after a successful message sending operation.
type MessageResponse struct {
	// Message contains the full details of the newly created message, including user names.
	Message MessageWithUser `json:"message"`
	// Success indicates the outcome of the operation (true for successful creation).
	Success bool `json:"success"`
}

// ConversationsResponse contains the list of conversation summaries for the current user.
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

// MessagesResponse contains a paginated list of messages for a specific conversation.
type MessagesResponse struct {
	// Messages is the slice of messages returned for the current page.
	Messages []MessageWithUser `json:"messages"`
	// HasMore indicates if there are more pages of messages available.
	HasMore bool `json:"has_more"`
	// Total is the total count of messages in the conversation.
	Total int `json:"total"`
}