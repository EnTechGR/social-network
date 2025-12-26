package message

import (
	"database/sql"
	"fmt"

	"social-network/models"
)

// GetConversation retrieves paginated messages between two specified users.
//
// It performs a two-way check (A->B or B->A) to include all messages in the thread,
// joins user details for sender/receiver names, and orders by time descending.
// The result includes attached chat image metadata if available.
//
// Parameters:
//   - userID: The ID of the current authenticated user.
//   - otherUserID: The ID of the peer user in the conversation.
//   - limit: The maximum number of messages to return per page.
//   - offset: The number of records to skip (for pagination).
//
// Returns:
//   - []models.MessageWithUser: A slice of messages with user details.
//   - error: An error if the query or scanning fails.
func (r *MessageRepository) GetConversation(userID, otherUserID string, limit, offset int) ([]models.MessageWithUser, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.DB.Query(`
		SELECT 
			m.message_id,
			m.sender_id,
			s.username AS sender_name,
			m.receiver_id,
			rec.username AS receiver_name,
			m.content,
			m.created_at,
			m.is_read
		FROM messages m
		JOIN user s ON m.sender_id = s.user_id
		JOIN user rec ON m.receiver_id = rec.user_id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, otherUserID, otherUserID, userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %v", err)
	}
	defer rows.Close()

	var messages []models.MessageWithUser
	for rows.Next() {
		var msg models.MessageWithUser
		err := rows.Scan(
			&msg.MessageID,
			&msg.SenderID,
			&msg.SenderNickname,
			&msg.ReceiverID,
			&msg.ReceiverName,
			&msg.Content,
			&msg.CreatedAt,
			&msg.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		// Fetch images for the current message
		images, err := r.GetChatImagesByMessageID(msg.MessageID)
		if err != nil {
			// Deciding whether to log and continue or return error is application specific.
			// Returning error here ensures data integrity.
			return nil, fmt.Errorf("failed to retrieve chat images for message %s: %w", msg.MessageID, err)
		}

		if len(images) > 0 {
			// Assuming 1:1 message:image relationship based on your table design
			msg.Image = images[0]
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %v", err)
	}

	return messages, nil
}

// GetConversationCount returns the total number of messages exchanged between two users.
//
// Parameters:
//   - userID: The ID of the first user.
//   - otherUserID: The ID of the second user.
//
// Returns:
//   - int: The total count of messages.
//   - error: An error if the query fails.
func (r *MessageRepository) GetConversationCount(userID, otherUserID string) (int, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) 
		   OR (sender_id = ? AND receiver_id = ?)
	`, userID, otherUserID, otherUserID, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get conversation count: %v", err)
	}

	return count, nil
}

// GetUnreadCount returns the number of unread messages sent by a specific user (fromUserID)
// to the current user (userID).
//
// Parameters:
//   - userID: The ID of the message recipient (current user).
//   - fromUserID: The ID of the message sender (peer user).
//
// Returns:
//   - int: The count of unread messages.
//   - error: An error if the query fails.
func (r *MessageRepository) GetUnreadCount(userID, fromUserID string) (int, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = ? AND sender_id = ? AND is_read = 0
	`, userID, fromUserID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %v", err)
	}

	return count, nil
}

// GetTotalUnreadCount returns the overall total number of unread messages for a given user.
//
// Parameters:
//   - userID: The ID of the user whose unread count is being checked.
//
// Returns:
//   - int: The total count of all unread messages.
//   - error: An error if the query fails.
func (r *MessageRepository) GetTotalUnreadCount(userID string) (int, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = ? AND is_read = 0
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get total unread count: %v", err)
	}

	return count, nil
}

// GetConversations retrieves a summary list of all active conversations for a user.
//
// It uses Common Table Expressions (CTEs) and subqueries to efficiently:
// 1. Identify all unique peer users (other_user_id).
// 2. Fetch the content and time of the last message for each thread.
// 3. Calculate the unread count for each thread.
// The list is ordered by the most recent message time.
//
// Parameters:
//   - userID: The ID of the user whose conversations are being retrieved.
//
// Returns:
//   - []models.Conversation: A slice of conversation summaries.
//   - error: An error if the query or scanning fails.
func (r *MessageRepository) GetConversations(userID string) ([]models.Conversation, error) {
	// Query to get conversation summaries
	rows, err := r.DB.Query(`
		WITH user_conversations AS (
			SELECT DISTINCT
				CASE 
					WHEN m.sender_id = ? THEN m.receiver_id
					ELSE m.sender_id
				END AS other_user_id
			FROM messages m
			WHERE m.sender_id = ? OR m.receiver_id = ?
		)
		SELECT 
			uc.other_user_id,
			u.username,
			(
				SELECT content
				FROM messages m2
				WHERE (m2.sender_id = ? AND m2.receiver_id = uc.other_user_id)
				   OR (m2.sender_id = uc.other_user_id AND m2.receiver_id = ?)
				ORDER BY m2.created_at DESC
				LIMIT 1
			) AS last_message,
			(
				SELECT created_at
				FROM messages m2
				WHERE (m2.sender_id = ? AND m2.receiver_id = uc.other_user_id)
				   OR (m2.sender_id = uc.other_user_id AND m2.receiver_id = ?)
				ORDER BY m2.created_at DESC
				LIMIT 1
			) AS last_message_time,
			(
				SELECT COUNT(*)
				FROM messages m3
				WHERE m3.receiver_id = ? AND m3.sender_id = uc.other_user_id AND m3.is_read = 0
			) AS unread_count
		FROM user_conversations uc
		JOIN user u ON u.user_id = uc.other_user_id
		ORDER BY last_message_time DESC
	`, userID, userID, userID, userID, userID, userID, userID, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %v", err)
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var lastMessage sql.NullString

		err := rows.Scan(
			&conv.UserID,
			&conv.Nickname,
			&lastMessage,
			&conv.LastMessageTime,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %v", err)
		}

		if lastMessage.Valid {
			conv.LastMessage = lastMessage.String
		} else {
			conv.LastMessage = ""
		}

		conv.IsOnline = false

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %v", err)
	}

	return conversations, nil
}

// GetUsersWithoutConversation retrieves users who have NOT yet exchanged messages with the current user.
//
// This is used, for example, to populate a list of potential new chat partners.
//
// Parameters:
//   - userID: The ID of the current user.
//
// Returns:
//   - []models.User: A slice of user models for users without a conversation history.
//   - error: An error if the query fails.
func (r *MessageRepository) GetUsersWithoutConversation(userID string) ([]models.User, error) {
	rows, err := r.DB.Query(`
		SELECT u.user_id, u.username, u.email, u.first_name, u.last_name, u.age, u.gender, u.created_at
		FROM user u
		WHERE u.user_id != ?
		AND u.user_id NOT IN (
			SELECT DISTINCT 
				CASE 
					WHEN sender_id = ? THEN receiver_id
					ELSE sender_id
				END
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
		)
		ORDER BY u.username ASC
	`, userID, userID, userID, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get users without conversation: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID,
			&u.Nickname,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.DateOfBirth,
			&u.Gender,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// GetAllUsers retrieves all user profiles in the system, excluding the current user.
// Used for displaying a full list of potential chat contacts.
//
// Parameters:
//   - currentUserID: The ID of the user to exclude from the list.
//
// Returns:
//   - []models.User: A slice of all other user profiles.
//   - error: An error if the query fails.
func (r *MessageRepository) GetAllUsers(currentUserID string) ([]models.User, error) {
	rows, err := r.DB.Query(`
		SELECT user_id, username, email, first_name, last_name, age, gender, created_at
		FROM user
		WHERE user_id != ?
		ORDER BY username ASC
	`, currentUserID)

	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID,
			&u.Nickname,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.DateOfBirth,
			&u.Gender,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}
	fmt.Println("Users retrieved from DB:")
	for _, u := range users {
		fmt.Printf("%+v\n", u)
	}

	return users, nil
}
