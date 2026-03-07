package models

import "time"

// GroupMember represents a user's membership in a group
// swagger:model GroupMember
type GroupMember struct {
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The ID of the user who is a member
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The timestamp when the user joined the group
	// example: 2025-12-29T18:00:00Z
	JoinedAt time.Time `json:"joined_at"`
}

// GroupMemberWithUser represents a group member with user details
// swagger:model GroupMemberWithUser
type GroupMemberWithUser struct {
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The ID of the user
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The display name of the user
	// example: johndoe_99
	Nickname string `json:"nickname"`
	// The user's email address
	// example: john.doe@example.com
	Email string `json:"email"`
	// The user's first name
	// example: John
	FirstName string `json:"first_name"`
	// The user's last name
	// example: Doe
	LastName string `json:"last_name"`
	// Whether this user is the owner of the group
	// example: false
	IsOwner bool `json:"is_owner"`
	// The timestamp when the user joined the group
	// example: 2025-12-29T18:00:00Z
	JoinedAt time.Time `json:"joined_at"`
}