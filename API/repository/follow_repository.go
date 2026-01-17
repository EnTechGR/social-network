package repository

import (
	"database/sql"
	"time"

	"social-network/models"
)

// FollowRepository handles follow relationship operations.
type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) createFollowWithStatus(followerID, followeeID, status string) (*models.FollowRelationship, error) {
	if followerID == followeeID {
		return nil, ErrFollowSelf
	}

	var existingStatus string
	err := r.db.QueryRow(`
		SELECT status 
		FROM follow_relationships 
		WHERE follower_id = ? AND followee_id = ?
	`, followerID, followeeID).Scan(&existingStatus)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		if existingStatus == "blocked" {
			return nil, ErrFollowBlocked
		}
		return nil, ErrFollowAlreadyExists
	}

	now := time.Now()
	_, err = r.db.Exec(`
		INSERT INTO follow_relationships (follower_id, followee_id, status, created_at) 
		VALUES (?, ?, ?, ?)
	`, followerID, followeeID, status, now)
	if err != nil {
		return nil, err
	}

	return &models.FollowRelationship{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     status,
		CreatedAt:  now,
	}, nil
}

// CreateAcceptedFollow creates an accepted follow relationship.
// This is used when a public user follows another public user.
func (r *FollowRepository) CreateAcceptedFollow(followerID, followeeID string) (*models.FollowRelationship, error) {
	return r.createFollowWithStatus(followerID, followeeID, "accepted")
}

// CreateFollowRequest creates a pending follow request.
func (r *FollowRepository) CreateFollowRequest(followerID, followeeID string) (*models.FollowRelationship, error) {
	return r.createFollowWithStatus(followerID, followeeID, "pending")
}

// GetByFolloweeAndStatus returns follow relationships for a followee with a specific status.
func (r *FollowRepository) GetByFolloweeAndStatus(followeeID, status string) ([]models.FollowRelationship, error) {
	rows, err := r.db.Query(`
		SELECT follower_id, followee_id, status, created_at, updated_at
		FROM follow_relationships
		WHERE followee_id = ? AND status = ?
		ORDER BY created_at DESC
	`, followeeID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []models.FollowRelationship
	for rows.Next() {
		var relationship models.FollowRelationship
		var updatedAt sql.NullTime
		if err := rows.Scan(&relationship.FollowerID, &relationship.FolloweeID, &relationship.Status, &relationship.CreatedAt, &updatedAt); err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			relationship.UpdatedAt = &updatedAt.Time
		}
		relationships = append(relationships, relationship)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return relationships, nil
}

// AcceptFollowRequest marks a pending follow request as accepted.
func (r *FollowRepository) AcceptFollowRequest(followerID, followeeID string) error {
	result, err := r.db.Exec(`
		UPDATE follow_relationships 
		SET status = 'accepted', updated_at = ?
		WHERE follower_id = ? AND followee_id = ? AND status = 'pending'
	`, time.Now(), followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// DeclineFollowRequest deletes a pending follow request.
func (r *FollowRepository) DeclineFollowRequest(followerID, followeeID string) error {
	result, err := r.db.Exec(`
		DELETE FROM follow_relationships 
		WHERE follower_id = ? AND followee_id = ? AND status = 'pending'
	`, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// DeleteFollow removes an accepted follow relationship.
func (r *FollowRepository) DeleteFollow(followerID, followeeID string) error {
	result, err := r.db.Exec(`
		DELETE FROM follow_relationships 
		WHERE follower_id = ? AND followee_id = ? AND status = 'accepted'
	`, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// RemoveFollower removes an accepted follow relationship where the current user is the followee.
func (r *FollowRepository) RemoveFollower(followeeID, followerID string) error {
	result, err := r.db.Exec(`
		DELETE FROM follow_relationships 
		WHERE follower_id = ? AND followee_id = ? AND status = 'accepted'
	`, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}
