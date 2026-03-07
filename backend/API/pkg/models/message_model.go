package models

import "time"

// Message represents a private message record
// swagger:model Message
type Message struct {
	// The unique identifier for the message (UUID)
	// example: m1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	MessageID string `json:"message_id"`
	// The ID of the user who sent the message
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	SenderID string `json:"sender_id"`
	// The ID of the user intended to receive the message
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	ReceiverID string `json:"receiver_id"`
	// The text body of the message
	// example: Hey, did you see the new documentation?
	Content string `json:"content"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// Indicates if the receiver has read the message
	// example: false
	IsRead bool `json:"is_read"`
	// Optional image attachment metadata
	Image *ChatImage `json:"image,omitempty"`
}

// ChatImage represents metadata for an image attached to a chat
// swagger:model ChatImage
type ChatImage struct {
	// example: img_987654321
	ImageID string `json:"image_id"`
	// example: m1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	MessageID string `json:"message_id"`
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// Internal server filename
	// example: 20251229_chat_xyz.png
	Filename string `json:"filename"`
	// Original filename provided by the user
	// example: screenshot.png
	OriginalFilename string `json:"original_filename"`
	// example: /uploads/chat_images/20251229_chat_xyz.png
	FilePath string `json:"file_path"`
	// example: /uploads/chat_images/thumbnails/20251229_chat_xyz.png
	ThumbnailPath string `json:"thumbnail_path"`
	// example: 1048576
	FileSize int64 `json:"file_size"`
	// example: image/png
	MimeType string `json:"mime_type"`
	// example: 1920
	Width int `json:"width"`
	// example: 1080
	Height int `json:"height"`
	// example: 2025-12-29T18:00:00Z
	UploadedAt time.Time `json:"uploaded_at"`
}

// MessageWithUser is the payload for API responses including nicknames
// swagger:model MessageWithUser
type MessageWithUser struct {
	// example: m1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	MessageID string `json:"message_id"`
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	SenderID string `json:"sender_id"`
	// example: gopher_king
	SenderNickname string `json:"sender_name"`
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	ReceiverID string `json:"receiver_id"`
	// example: tech_lead
	ReceiverName string `json:"receiver_name"`
	// example: Hey, did you see the new documentation?
	Content string `json:"content"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: true
	IsRead bool       `json:"is_read"`
	Image  *ChatImage `json:"image,omitempty"`
}

// CreateMessageRequest defines the body for sending a new message
// swagger:model CreateMessageRequest
type CreateMessageRequest struct {
	// ID of the recipient
	// required: true
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	ReceiverID string `json:"receiver_id" binding:"required"`
	// Message content
	// required: true
	// example: Hello there!
	Content string `json:"content" binding:"required"`
}

// Conversation summarizes a chat thread with another user
// swagger:model Conversation
type Conversation struct {
	// The peer user's ID
	// example: f2g3h4i5-j6k7-l8m9-n0o1-p2q3r4s5t6u7
	UserID string `json:"user_id"`
	// The peer user's nickname
	// example: tech_lead
	Nickname string `json:"nickname"`
	// The peer user's avatar metadata when available
	Avatar *AvatarInfo `json:"avatar,omitempty"`
	// The last message sent in this thread
	// example: Sounds good, see you then.
	LastMessage string `json:"last_message"`
	// example: 2025-12-29T18:05:00Z
	LastMessageTime time.Time `json:"last_message_time"`
	// Count of unread messages for the current user
	// example: 3
	UnreadCount int `json:"unread_count"`
	// Real-time online status of the peer user
	// example: true
	IsOnline bool `json:"is_online"`
}

// MessageResponse is the envelope for a single message result
// swagger:model MessageResponse
type MessageResponse struct {
	Message MessageWithUser `json:"message"`
	// example: true
	Success bool `json:"success"`
}

// ConversationsResponse is the envelope for the conversation list
// swagger:model ConversationsResponse
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

// MessagesResponse is the envelope for a conversation history
// swagger:model MessagesResponse
type MessagesResponse struct {
	Messages []MessageWithUser `json:"messages"`
	// True if more messages can be loaded (pagination)
	// example: true
	HasMore bool `json:"has_more"`
	// Total messages in this specific conversation
	// example: 150
	Total int `json:"total"`
}

// GroupMessage represents a message sent in a group chat room.
// swagger:model GroupMessage
type GroupMessage struct {
	// example: 31b7d4e4-6ef9-42cc-b8af-9ad4a8ecf63f
	MessageID string `json:"message_id"`
	// example: e01a5a0b-668b-4ae8-95e3-8ca482d900f9
	GroupID string `json:"group_id"`
	// example: 0b1fb3bb-8db0-4f1d-9abf-b0e49d2f2e60
	SenderID string `json:"sender_id"`
	// example: green_fox
	SenderNickname string `json:"sender_name"`
	// example: hello team 👋
	Content string `json:"content"`
	// example: 2026-02-21T12:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// CreateGroupMessageRequest defines the body for sending a group chat message.
// swagger:model CreateGroupMessageRequest
type CreateGroupMessageRequest struct {
	// Message content
	// required: true
	// example: Hi everyone!
	Content string `json:"content" binding:"required"`
}

// GroupMessagesResponse is the envelope for group chat history.
// swagger:model GroupMessagesResponse
type GroupMessagesResponse struct {
	Messages []GroupMessage `json:"messages"`
	HasMore  bool           `json:"has_more"`
	Total    int            `json:"total"`
}
