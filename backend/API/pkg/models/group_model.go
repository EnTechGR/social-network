package models

import "time"

// Group represents a user-created group in its raw database format
// swagger:model Group
type Group struct {
	// The unique identifier for the group (UUID)
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the user who created the group (owner)
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	OwnerID string `json:"owner_id"`
	// The title of the group
	// example: Photography Enthusiasts
	Title string `json:"title"`
	// The description of the group
	// example: A group for sharing photography tips and showcasing work
	Description *string `json:"description,omitempty"`
	// The timestamp when the group was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// GroupWithOwner represents a group with owner information for the UI
// swagger:model GroupWithOwner
type GroupWithOwner struct {
	// The unique identifier for the group (UUID)
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the owner
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	OwnerID string `json:"owner_id"`
	// The display name of the owner
	// example: photo_master
	OwnerNickname string `json:"owner_nickname"`
	// The title of the group
	// example: Photography Enthusiasts
	Title string `json:"title"`
	// The description of the group
	// example: A group for sharing photography tips and showcasing work
	Description *string `json:"description,omitempty"`
	// The number of members in the group
	// example: 42
	MemberCount int `json:"member_count"`
	// The timestamp when the group was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// GroupWithMembership represents a group with current user's membership status
// swagger:model GroupWithMembership
type GroupWithMembership struct {
	// The unique identifier for the group (UUID)
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the owner
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	OwnerID string `json:"owner_id"`
	// The display name of the owner
	// example: photo_master
	OwnerNickname string `json:"owner_nickname"`
	// The title of the group
	// example: Photography Enthusiasts
	Title string `json:"title"`
	// The description of the group
	// example: A group for sharing photography tips and showcasing work
	Description *string `json:"description,omitempty"`
	// The number of members in the group
	// example: 42
	MemberCount int `json:"member_count"`
	// Whether the current user is a member
	// example: true
	IsMember bool `json:"is_member"`
	// Whether the current user is the owner
	// example: false
	IsOwner bool `json:"is_owner"`
	// Whether the current user has a pending invite
	// example: false
	HasPendingInvite bool `json:"has_pending_invite"`
	// Whether the current user has a pending join request
	// example: false
	HasPendingRequest bool `json:"has_pending_request"`
	// The timestamp when the group was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}