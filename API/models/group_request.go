package models

import "time"

// GroupJoinRequest represents a user's request to join a group
// swagger:model GroupJoinRequest
type GroupJoinRequest struct {
	// The unique identifier for the request (UUID)
	// example: r1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The ID of the user requesting to join
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The status of the request: pending, approved, denied
	// example: pending
	Status string `json:"status"`
	// The timestamp when the request was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the request was responded to
	// example: 2025-12-29T19:30:00Z
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// GroupJoinRequestWithDetails represents a join request with user and group details
// swagger:model GroupJoinRequestWithDetails
type GroupJoinRequestWithDetails struct {
	// The unique identifier for the request (UUID)
	// example: r1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The title of the group
	// example: Photography Enthusiasts
	GroupTitle string `json:"group_title"`
	// The ID of the user requesting to join
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The display name of the user
	// example: johndoe_99
	UserNickname string `json:"user_nickname"`
	// The user's first name
	// example: John
	FirstName string `json:"first_name"`
	// The user's last name
	// example: Doe
	LastName string `json:"last_name"`
	// The status of the request: pending, approved, denied
	// example: pending
	Status string `json:"status"`
	// The timestamp when the request was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the request was responded to
	// example: 2025-12-29T19:30:00Z
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}