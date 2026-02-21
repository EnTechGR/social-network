package message

import (
	"database/sql"
	"fmt"
	"time"
	"unicode/utf8"

	"social-network/models"
	"social-network/utils"
)

// Create sends a new message and persists it to the database.
//
// It performs basic content validation (length and emptiness) before insertion.
//
// Parameters:
//   - senderID: The ID of the user sending the message.
//   - receiverID: The ID of the intended recipient.
//   - content: The text body of the message.
//
// Returns:
//   - *models.Message: The newly created message object.
//   - error: An error if validation fails or database insertion fails.
func (r *MessageRepository) Create(senderID, receiverID, content string) (*models.Message, error) {
	// Validate that sender and receiver are different
	if senderID == receiverID {
		return nil, fmt.Errorf("cannot send message to yourself")
	}
	// Validate content
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	// Validate content length
	if utf8.RuneCountInString(content) > 1000 {
		return nil, fmt.Errorf("message content cannot exceed 1000 characters")
	}
	// Generate unique message ID and timestamp
	messageID := utils.GenerateUUID()
	createdAt := time.Now()

	// Insert the new message into the database
	_, err := r.DB.Exec(`
		INSERT INTO messages (message_id, sender_id, receiver_id, content, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, ?)
	`, messageID, senderID, receiverID, content, createdAt, false)

	// Handle insertion errors
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	// Return the created message object
	return &models.Message{
		MessageID:  messageID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  createdAt,
		IsRead:     false,
	}, nil
}

// GetByID retrieves a single message record by its unique identifier.
//
// It also fetches associated image metadata and includes it in the returned model.
//
// Parameters:
//   - messageID: The unique ID of the message to retrieve.
//
// Returns:
//   - *models.Message: The message object, including optional image metadata.
//   - error: An error if the message is not found or the query fails.
func (r *MessageRepository) GetByID(messageID string) (*models.Message, error) {
	// Retrieve the message record from the database
	var msg models.Message

	// Query the message details
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

	// Handle errors during retrieval
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %v", err)
	}

	// Retrieve associated chat images, if any
	images, err := r.GetChatImagesByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chat images for message %s: %w", messageID, err)
	}

	// Attach the first image found, if available
	if len(images) > 0 {
		msg.Image = images[0]
	}

	// Return the complete message object
	return &msg, nil
}

// MarkAsRead updates the 'is_read' status for a single message to true (1).
//
// It first verifies that the user attempting the update is the intended receiver.
//
// Parameters:
//   - messageID: The ID of the message to be marked.
//   - userID: The ID of the user performing the action (must be the receiver).
//
// Returns:
//   - error: An error if the message is not found or the user is unauthorized.
func (r *MessageRepository) MarkAsRead(messageID, userID string) error {
	// Verify the user is the receiver of the message
	var receiverID string
	err := r.DB.QueryRow(`
		SELECT receiver_id FROM messages WHERE message_id = ?
	`, messageID).Scan(&receiverID)

	// Handle verification errors
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to verify message: %v", err)
	}

	// Check authorization
	if receiverID != userID {
		return fmt.Errorf("unauthorized: only receiver can mark message as read")
	}

	_, err = r.DB.Exec(`
		UPDATE messages
		SET is_read = 1
		WHERE message_id = ?
	`, messageID)

	// Handle update errors
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %v", err)
	}

	return nil
}

// MarkConversationAsRead marks all unread messages sent by 'otherUserID' to 'userID' as read (1).
//
// Parameters:
//   - userID: The ID of the user who received the messages (the one marking them as read).
//   - otherUserID: The ID of the peer user who sent the messages.
//
// Returns:
//   - error: An error if the update query fails.
func (r *MessageRepository) MarkConversationAsRead(userID, otherUserID string) error {
	// Update all unread messages in the conversation to read
	_, err := r.DB.Exec(`
		UPDATE messages
		SET is_read = 1
		WHERE receiver_id = ? AND sender_id = ? AND is_read = 0
	`, userID, otherUserID)
	// Handle update errors
	if err != nil {
		return fmt.Errorf("failed to mark conversation as read: %v", err)
	}

	return nil
}

// Delete removes a message permanently from the database.
//
// It first verifies that the user requesting the deletion is the original sender.
//
// Parameters:
//   - messageID: The ID of the message to be deleted.
//   - userID: The ID of the user attempting to delete (must be the sender).
//
// Returns:
//   - error: An error if the message is not found or the user is unauthorized
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
	// Check authorization
	if senderID != userID {
		return fmt.Errorf("unauthorized: only sender can delete message")
	}
	// Perform the deletion
	_, err = r.DB.Exec(`
		DELETE FROM messages WHERE message_id = ?
	`, messageID)

	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}

	return nil
}
