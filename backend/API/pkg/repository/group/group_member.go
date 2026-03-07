package group

import (
	"database/sql"
	"social-network/pkg/models"
	"time"
)

type GroupMemberRepository struct {
	db *sql.DB
}

func NewGroupMemberRepository(db *sql.DB) *GroupMemberRepository {
	return &GroupMemberRepository{db: db}
}

// AddMember adds a user to a group
func (r *GroupMemberRepository) AddMember(groupID, userID string) error {
	_, err := r.db.Exec(
		`INSERT INTO group_members (group_id, user_id, joined_at) 
		VALUES (?, ?, ?)`,
		groupID, userID, time.Now(),
	)
	return err
}

// RemoveMember removes a user from a group
func (r *GroupMemberRepository) RemoveMember(groupID, userID string) error {
	result, err := r.db.Exec(
		`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetGroupMembers retrieves all members of a group with user details
func (r *GroupMemberRepository) GetGroupMembers(groupID string) ([]models.GroupMemberWithUser, error) {
	rows, err := r.db.Query(
		`SELECT 
			gm.group_id, gm.user_id, u.nickname, u.email, u.first_name, u.last_name,
			(g.owner_id = gm.user_id) as is_owner, gm.joined_at
		FROM group_members gm
		INNER JOIN user u ON gm.user_id = u.user_id
		INNER JOIN groups g ON gm.group_id = g.group_id
		WHERE gm.group_id = ?
		ORDER BY is_owner DESC, gm.joined_at ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMemberWithUser
	for rows.Next() {
		var m models.GroupMemberWithUser
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Nickname, &m.Email, &m.FirstName, &m.LastName, &m.IsOwner, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetUserGroupIDs retrieves all group IDs a user is a member of
func (r *GroupMemberRepository) GetUserGroupIDs(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT group_id FROM group_members WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, rows.Err()
}

// IsMember checks if a user is a member of a group
func (r *GroupMemberRepository) IsMember(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMemberCount returns the number of members in a group
func (r *GroupMemberRepository) GetMemberCount(groupID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_members WHERE group_id = ?`,
		groupID,
	).Scan(&count)
	return count, err
}