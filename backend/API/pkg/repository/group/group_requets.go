package group

import (
	"database/sql"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"time"
)

type GroupJoinRequestRepository struct {
	db *sql.DB
}

func NewGroupJoinRequestRepository(db *sql.DB) *GroupJoinRequestRepository {
	return &GroupJoinRequestRepository{db: db}
}

// Create creates a new join request
func (r *GroupJoinRequestRepository) Create(groupID, userID string) (*models.GroupJoinRequest, error) {
	request := models.GroupJoinRequest{
		ID:        utils.GenerateUUID(),
		GroupID:   groupID,
		UserID:    userID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	_, err := r.db.Exec(
		`INSERT INTO group_join_requests (request_id, group_id, user_id, status, created_at) 
		VALUES (?, ?, ?, ?, ?)`,
		request.ID, request.GroupID, request.UserID, request.Status, request.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

// GetByID retrieves a join request by ID
func (r *GroupJoinRequestRepository) GetByID(requestID string) (*models.GroupJoinRequest, error) {
	var req models.GroupJoinRequest
	err := r.db.QueryRow(
		`SELECT request_id, group_id, user_id, status, created_at, responded_at 
		FROM group_join_requests WHERE request_id = ?`,
		requestID,
	).Scan(&req.ID, &req.GroupID, &req.UserID, &req.Status, &req.CreatedAt, &req.RespondedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// GetPendingRequestsForGroup retrieves all pending join requests for a group
func (r *GroupJoinRequestRepository) GetPendingRequestsForGroup(groupID string) ([]models.GroupJoinRequestWithDetails, error) {
	rows, err := r.db.Query(
		`SELECT 
			gjr.request_id, gjr.group_id, g.title, gjr.user_id, u.nickname, 
			u.first_name, u.last_name, gjr.status, gjr.created_at, gjr.responded_at
		FROM group_join_requests gjr
		INNER JOIN groups g ON gjr.group_id = g.group_id
		INNER JOIN user u ON gjr.user_id = u.user_id
		WHERE gjr.group_id = ? AND gjr.status = 'pending'
		ORDER BY gjr.created_at ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.GroupJoinRequestWithDetails
	for rows.Next() {
		var r models.GroupJoinRequestWithDetails
		if err := rows.Scan(&r.ID, &r.GroupID, &r.GroupTitle, &r.UserID, &r.UserNickname, &r.FirstName, &r.LastName, &r.Status, &r.CreatedAt, &r.RespondedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

// GetUserRequests retrieves all join requests made by a user
func (r *GroupJoinRequestRepository) GetUserRequests(userID string) ([]models.GroupJoinRequestWithDetails, error) {
	rows, err := r.db.Query(
		`SELECT 
			gjr.request_id, gjr.group_id, g.title, gjr.user_id, u.nickname, 
			u.first_name, u.last_name, gjr.status, gjr.created_at, gjr.responded_at
		FROM group_join_requests gjr
		INNER JOIN groups g ON gjr.group_id = g.group_id
		INNER JOIN user u ON gjr.user_id = u.user_id
		WHERE gjr.user_id = ?
		ORDER BY gjr.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.GroupJoinRequestWithDetails
	for rows.Next() {
		var r models.GroupJoinRequestWithDetails
		if err := rows.Scan(&r.ID, &r.GroupID, &r.GroupTitle, &r.UserID, &r.UserNickname, &r.FirstName, &r.LastName, &r.Status, &r.CreatedAt, &r.RespondedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

// UpdateStatus updates the status of a join request (approve/deny)
func (r *GroupJoinRequestRepository) UpdateStatus(requestID, status string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var groupID, userID string
	err = tx.QueryRow(
		`SELECT group_id, user_id FROM group_join_requests WHERE request_id = ?`,
		requestID,
	).Scan(&groupID, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}

	// Keep the newest decision by removing any previous rows with the same
	// (group_id, user_id, status) tuple. This avoids UNIQUE conflicts when a
	// user leaves then gets approved/denied again later.
	if _, err := tx.Exec(
		`DELETE FROM group_join_requests
		 WHERE group_id = ? AND user_id = ? AND status = ? AND request_id <> ?`,
		groupID, userID, status, requestID,
	); err != nil {
		return err
	}

	now := time.Now()
	result, err := tx.Exec(
		`UPDATE group_join_requests SET status = ?, responded_at = ? WHERE request_id = ?`,
		status, now, requestID,
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

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Delete removes a join request
func (r *GroupJoinRequestRepository) Delete(requestID string) error {
	result, err := r.db.Exec(`DELETE FROM group_join_requests WHERE request_id = ?`, requestID)
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

// HasPendingRequest checks if a user has a pending join request for a group
func (r *GroupJoinRequestRepository) HasPendingRequest(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_join_requests 
		WHERE group_id = ? AND user_id = ? AND status = 'pending'`,
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
