package group

import (
	"database/sql"
	"social-network/models"
	"social-network/utils"
	"time"
)

type GroupInviteRepository struct {
	db *sql.DB
}

func NewGroupInviteRepository(db *sql.DB) *GroupInviteRepository {
	return &GroupInviteRepository{db: db}
}

// Create creates a new group invitation
func (r *GroupInviteRepository) Create(groupID, fromUserID, toUserID string) (*models.GroupInvite, error) {
	invite := models.GroupInvite{
		ID:         utils.GenerateUUID(),
		GroupID:    groupID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	_, err := r.db.Exec(
		`INSERT INTO group_invites (invite_id, group_id, from_user_id, to_user_id, status, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		invite.ID, invite.GroupID, invite.FromUserID, invite.ToUserID, invite.Status, invite.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

// GetByID retrieves an invite by ID
func (r *GroupInviteRepository) GetByID(inviteID string) (*models.GroupInvite, error) {
	var invite models.GroupInvite
	err := r.db.QueryRow(
		`SELECT invite_id, group_id, from_user_id, to_user_id, status, created_at, responded_at 
		FROM group_invites WHERE invite_id = ?`,
		inviteID,
	).Scan(&invite.ID, &invite.GroupID, &invite.FromUserID, &invite.ToUserID, &invite.Status, &invite.CreatedAt, &invite.RespondedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &invite, nil
}

// GetPendingInvitesForUser retrieves all pending invites for a user
func (r *GroupInviteRepository) GetPendingInvitesForUser(userID string) ([]models.GroupInviteWithDetails, error) {
	rows, err := r.db.Query(
		`SELECT 
			gi.invite_id, gi.group_id, g.title, gi.from_user_id, u1.nickname, 
			gi.to_user_id, u2.nickname, gi.status, gi.created_at, gi.responded_at
		FROM group_invites gi
		INNER JOIN groups g ON gi.group_id = g.group_id
		INNER JOIN user u1 ON gi.from_user_id = u1.user_id
		INNER JOIN user u2 ON gi.to_user_id = u2.user_id
		WHERE gi.to_user_id = ? AND gi.status = 'pending'
		ORDER BY gi.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []models.GroupInviteWithDetails
	for rows.Next() {
		var i models.GroupInviteWithDetails
		if err := rows.Scan(&i.ID, &i.GroupID, &i.GroupTitle, &i.FromUserID, &i.FromNickname, &i.ToUserID, &i.ToNickname, &i.Status, &i.CreatedAt, &i.RespondedAt); err != nil {
			return nil, err
		}
		invites = append(invites, i)
	}
	return invites, rows.Err()
}

// GetGroupInvites retrieves all invites for a specific group
func (r *GroupInviteRepository) GetGroupInvites(groupID string) ([]models.GroupInviteWithDetails, error) {
	rows, err := r.db.Query(
		`SELECT 
			gi.invite_id, gi.group_id, g.title, gi.from_user_id, u1.nickname, 
			gi.to_user_id, u2.nickname, gi.status, gi.created_at, gi.responded_at
		FROM group_invites gi
		INNER JOIN groups g ON gi.group_id = g.group_id
		INNER JOIN user u1 ON gi.from_user_id = u1.user_id
		INNER JOIN user u2 ON gi.to_user_id = u2.user_id
		WHERE gi.group_id = ?
		ORDER BY gi.created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []models.GroupInviteWithDetails
	for rows.Next() {
		var i models.GroupInviteWithDetails
		if err := rows.Scan(&i.ID, &i.GroupID, &i.GroupTitle, &i.FromUserID, &i.FromNickname, &i.ToUserID, &i.ToNickname, &i.Status, &i.CreatedAt, &i.RespondedAt); err != nil {
			return nil, err
		}
		invites = append(invites, i)
	}
	return invites, rows.Err()
}

// UpdateStatus updates the status of an invite (accept/decline)
func (r *GroupInviteRepository) UpdateStatus(inviteID, status string) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE group_invites SET status = ?, responded_at = ? WHERE invite_id = ?`,
		status, now, inviteID,
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

// Delete removes an invite
func (r *GroupInviteRepository) Delete(inviteID string) error {
	result, err := r.db.Exec(`DELETE FROM group_invites WHERE invite_id = ?`, inviteID)
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

// HasPendingInvite checks if a user has a pending invite to a group
func (r *GroupInviteRepository) HasPendingInvite(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_invites 
		WHERE group_id = ? AND to_user_id = ? AND status = 'pending'`,
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}