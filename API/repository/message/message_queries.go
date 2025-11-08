package message

import (
	"database/sql"
	"fmt"

	"forum/models"
)

// GetConversation retrieves messages between two users with pagination
// Returns the last 'limit' messages, ordered by most recent first
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
			&msg.SenderName,
			&msg.ReceiverID,
			&msg.ReceiverName,
			&msg.Content,
			&msg.CreatedAt,
			&msg.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %v", err)
	}

	return messages, nil
}

// GetConversationCount returns total message count between two users
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

// GetUnreadCount returns the number of unread messages for a user from another user
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

// GetTotalUnreadCount returns total unread messages for a user
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

// GetConversations retrieves all conversations for a user
// Ordered by last message time (most recent first)
func (r *MessageRepository) GetConversations(userID string) ([]models.Conversation, error) {
	// ✅ FIXED: Use a subquery to get the other_user_id first, then join to get details
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
			&conv.Username,
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

		// Note: IsOnline will be set by the WebSocket manager
		conv.IsOnline = false

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %v", err)
	}

	return conversations, nil
}

// GetUsersWithoutConversation retrieves users who don't have a conversation with the current user
// Useful for showing "new chat" options
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
			&u.Username,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.Age,
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

// GetAllUsers retrieves all users except the current user (for chat list)
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
			&u.Username,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.Age,
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
