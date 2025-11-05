package message

import (
	"database/sql"
	"fmt"
	"time"

	"forum/models"
	"forum/utils"
)

// Create sends a new message
func (r *MessageRepository) Create(senderID, receiverID, content string) (*models.Message, error) {
	// Validate that sender and receiver are different
	if senderID == receiverID {
		return nil, fmt.Errorf("cannot send message to yourself")
	}

	// Validate content
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	if len(content) > 1000 {
		return nil, fmt.Errorf("message content cannot exceed 1000 characters")
	}

	messageID := utils.GenerateUUID()
	createdAt := time.Now()

	_, err := r.DB.Exec(`
		INSERT INTO messages (message_id, sender_id, receiver_id, content, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, ?)
	`, messageID, senderID, receiverID, content, createdAt, false)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	return &models.Message{
		MessageID:  messageID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  createdAt,
		IsRead:     false,
	}, nil
}

// GetByID retrieves a message by its ID
func (r *MessageRepository) GetByID(messageID string) (*models.Message, error) {
	var msg models.Message

	err := r.DB.QueryRow(`
		SELECT message_id, sender_id, receiver_id, content, created_at, is_read
		FROM messages
		WHERE message_id = ?
	`, messageID).Scan(
		&msg.MessageID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.IsRead,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %v", err)
	}

	return &msg, nil
}

// MarkAsRead marks a message as read
func (r *MessageRepository) MarkAsRead(messageID, userID string) error {
	// Verify the user is the receiver of the message
	var receiverID string
	err := r.DB.QueryRow(`
		SELECT receiver_id FROM messages WHERE message_id = ?
	`, messageID).Scan(&receiverID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to verify message: %v", err)
	}

	if receiverID != userID {
		return fmt.Errorf("unauthorized: only receiver can mark message as read")
	}

	_, err = r.DB.Exec(`
		UPDATE messages
		SET is_read = 1
		WHERE message_id = ?
	`, messageID)

	if err != nil {
		return fmt.Errorf("failed to mark message as read: %v", err)
	}

	return nil
}

// MarkConversationAsRead marks all messages in a conversation as read
func (r *MessageRepository) MarkConversationAsRead(userID, otherUserID string) error {
	_, err := r.DB.Exec(`
		UPDATE messages
		SET is_read = 1
		WHERE receiver_id = ? AND sender_id = ? AND is_read = 0
	`, userID, otherUserID)

	if err != nil {
		return fmt.Errorf("failed to mark conversation as read: %v", err)
	}

	return nil
}

// Delete removes a message (only sender can delete)
func (r *MessageRepository) Delete(messageID, userID string) error {
	// Verify the user is the sender of the message
	var senderID string
	err := r.DB.QueryRow(`
		SELECT sender_id FROM messages WHERE message_id = ?
	`, messageID).Scan(&senderID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to verify message: %v", err)
	}

	if senderID != userID {
		return fmt.Errorf("unauthorized: only sender can delete message")
	}

	_, err = r.DB.Exec(`
		DELETE FROM messages WHERE message_id = ?
	`, messageID)

	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}

	return nil
}