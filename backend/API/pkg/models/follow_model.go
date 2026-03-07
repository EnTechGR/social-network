package models

import "time"

// FollowRelationship represents a follower-to-followee relationship.
// swagger:model FollowRelationship
type FollowRelationship struct {
	// The ID of the user who follows.
	FollowerID string `json:"follower_id"`
	// The ID of the user being followed.
	FolloweeID string `json:"followee_id"`
	// Relationship status: pending, accepted, blocked.
	Status string `json:"status"`
	// When the relationship was created.
	CreatedAt time.Time `json:"created_at"`
	// When the relationship was last updated.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
