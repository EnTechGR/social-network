package models

import "time"

// Message represents a private message between two users
type Message struct {
	MessageID  string     `json:"message_id"`
	SenderID   string     `json:"sender_id"`
	ReceiverID string     `json:"receiver_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	IsRead     bool       `json:"is_read"`
    // Add the image field (using a pointer so it can be nil/optional)
	Image      *ChatImage `json:"image,omitempty"` 
}

// MessageWithUser includes the sender's information for display
type MessageWithUser struct {
	MessageID    string     `json:"message_id"`
	SenderID     string     `json:"sender_id"`
	SenderName   string     `json:"sender_name"`
	ReceiverID   string     `json:"receiver_id"`
	ReceiverName string     `json:"receiver_name"`
	Content      string     `json:"content"`
	CreatedAt    time.Time  `json:"created_at"`
	IsRead       bool       `json:"is_read"`
    // Add the image field here as well
	Image        *ChatImage `json:"image,omitempty"` 
}

// CreateMessageRequest represents the request to send a new message
type CreateMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// Conversation represents a chat conversation with another user
type Conversation struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	LastMessage     string    `json:"last_message"`
	LastMessageTime time.Time `json:"last_message_time"`
	UnreadCount     int       `json:"unread_count"`
	IsOnline        bool      `json:"is_online"`
}

// MessageResponse is the response after sending a message
type MessageResponse struct {
	Message MessageWithUser `json:"message"`
	Success bool            `json:"success"`
}

// ConversationsResponse contains the list of conversations
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

// MessagesResponse contains paginated messages
type MessagesResponse struct {
	Messages []MessageWithUser `json:"messages"`
	HasMore  bool              `json:"has_more"`
	Total    int               `json:"total"`
}

// ChatImage represents an image sent through chat messages
type ChatImage struct {
	ImageID          string    `json:"image_id"`
	MessageID        string    `json:"message_id"`
	UserID           string    `json:"user_id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	FilePath         string    `json:"file_path"`
	ThumbnailPath    string    `json:"thumbnail_path"`
	FileSize         int64     `json:"file_size"`
	MimeType         string    `json:"mime_type"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	UploadedAt       time.Time `json:"uploaded_at"`
}