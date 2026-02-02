package models

import "time"

// GroupInvite represents an invitation to join a group
// swagger:model GroupInvite
type GroupInvite struct {
	// The unique identifier for the invite (UUID)
	// example: i1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The ID of the user who sent the invite
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	FromUserID string `json:"from_user_id"`
	// The ID of the user who received the invite
	// example: b2c3d4e5-f6g7-h8i9-j0k1-l2m3n4o5p6q7
	ToUserID string `json:"to_user_id"`
	// The status of the invite: pending, accepted, declined
	// example: pending
	Status string `json:"status"`
	// The timestamp when the invite was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the invite was responded to
	// example: 2025-12-29T19:30:00Z
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// GroupInviteWithDetails represents an invite with group and user details
// swagger:model GroupInviteWithDetails
type GroupInviteWithDetails struct {
	// The unique identifier for the invite (UUID)
	// example: i1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The title of the group
	// example: Photography Enthusiasts
	GroupTitle string `json:"group_title"`
	// The ID of the user who sent the invite
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	FromUserID string `json:"from_user_id"`
	// The display name of the user who sent the invite
	// example: photo_master
	FromNickname string `json:"from_nickname"`
	// The ID of the user who received the invite
	// example: b2c3d4e5-f6g7-h8i9-j0k1-l2m3n4o5p6q7
	ToUserID string `json:"to_user_id"`
	// The display name of the user who received the invite
	// example: johndoe_99
	ToNickname string `json:"to_nickname"`
	// The status of the invite: pending, accepted, declined
	// example: pending
	Status string `json:"status"`
	// The timestamp when the invite was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the invite was responded to
	// example: 2025-12-29T19:30:00Z
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}