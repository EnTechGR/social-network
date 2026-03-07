package message

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"social-network/pkg/models"
	"social-network/pkg/utils"
)

// CanUsersMessage returns true when two users have at least one accepted follow relation
// in either direction. Users with no accepted follow relation cannot start a direct chat.
func (r *MessageRepository) CanUsersMessage(userAID, userBID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follow_relationships
		WHERE status = 'accepted'
		  AND (
			(follower_id = ? AND followee_id = ?)
			OR
			(follower_id = ? AND followee_id = ?)
		  )
	`, userAID, userBID, userBID, userAID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate chat permission: %w", err)
	}
	return count > 0, nil
}

// CanDeliverMessageRealtime returns true when the receiver can get live websocket
// delivery for a direct message from senderID.
//
// Rule:
//   - receiver has a public profile, OR
//   - receiver follows sender with an accepted follow relationship.
func (r *MessageRepository) CanDeliverMessageRealtime(senderID, receiverID string) (bool, error) {
	var receiverIsPrivate bool
	err := r.DB.QueryRow(`
		SELECT is_private
		FROM user
		WHERE user_id = ?
	`, receiverID).Scan(&receiverIsPrivate)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("receiver not found")
		}
		return false, fmt.Errorf("failed to load receiver privacy: %w", err)
	}

	if !receiverIsPrivate {
		return true, nil
	}

	var count int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follow_relationships
		WHERE status = 'accepted'
		  AND follower_id = ?
		  AND followee_id = ?
	`, receiverID, senderID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate realtime chat delivery: %w", err)
	}

	return count > 0, nil
}

// IsGroupMember checks if a user is a member of a specific group.
func (r *MessageRepository) IsGroupMember(groupID, userID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to verify group membership: %w", err)
	}
	return count > 0, nil
}

// GetGroupMemberIDs returns all member user IDs for a group.
func (r *MessageRepository) GetGroupMemberIDs(groupID string) ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT user_id
		FROM group_members
		WHERE group_id = ?
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	memberIDs := make([]string, 0)
	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		memberIDs = append(memberIDs, memberID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating group members: %w", err)
	}

	return memberIDs, nil
}

// CreateGroupMessage persists a new group chat message.
func (r *MessageRepository) CreateGroupMessage(groupID, senderID, content string) (*models.GroupMessage, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if utf8.RuneCountInString(content) > 1000 {
		return nil, fmt.Errorf("message content cannot exceed 1000 characters")
	}

	messageID := utils.GenerateUUID()
	createdAt := time.Now()

	_, err := r.DB.Exec(`
		INSERT INTO group_messages (message_id, group_id, sender_id, content, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, messageID, groupID, senderID, content, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create group message: %w", err)
	}

	var senderNickname string
	if err := r.DB.QueryRow(`SELECT nickname FROM user WHERE user_id = ?`, senderID).Scan(&senderNickname); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sender not found")
		}
		return nil, fmt.Errorf("failed to load sender: %w", err)
	}

	return &models.GroupMessage{
		MessageID:      messageID,
		GroupID:        groupID,
		SenderID:       senderID,
		SenderNickname: senderNickname,
		Content:        content,
		CreatedAt:      createdAt,
	}, nil
}

// GetGroupMessages returns paginated group chat history ordered by newest first.
func (r *MessageRepository) GetGroupMessages(groupID string, limit, offset int) ([]models.GroupMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.DB.Query(`
		SELECT gm.message_id, gm.group_id, gm.sender_id, u.nickname, gm.content, gm.created_at
		FROM group_messages gm
		INNER JOIN user u ON u.user_id = gm.sender_id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at DESC
		LIMIT ? OFFSET ?
	`, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get group messages: %w", err)
	}
	defer rows.Close()

	messages := make([]models.GroupMessage, 0)
	for rows.Next() {
		var message models.GroupMessage
		if err := rows.Scan(
			&message.MessageID,
			&message.GroupID,
			&message.SenderID,
			&message.SenderNickname,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan group message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating group messages: %w", err)
	}

	return messages, nil
}

// GetGroupMessageCount returns total number of messages in a group chat.
func (r *MessageRepository) GetGroupMessageCount(groupID string) (int, error) {
	var count int
	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_messages
		WHERE group_id = ?
	`, groupID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count group messages: %w", err)
	}
	return count, nil
}
