package models

import "time"

// Reaction represents a user's emotional response (like/dislike) to content
// swagger:model Reaction
type Reaction struct {
	// The ID of the user who reacted
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The type of reaction.
	// 1 = Like, -1 = Dislike (or your specific mapping)
	// example: 1
	Type int `json:"reaction_type"`
	// The ID of the comment if the reaction is on a comment
	// example: 550e8400-e29b-41d4-a716-446655440000
	CommentID *string `json:"comment_id,omitempty"`
	// The ID of the post if the reaction is on a post
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID *string `json:"post_id,omitempty"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// ReactionWithUser represents a reaction including the participant's nickname
// swagger:model ReactionWithUser
type ReactionWithUser struct {
	// The ID of the user who reacted
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The display name of the user
	// example: gopher_fan
	Nickname string `json:"nickname"`
	// The type of reaction (e.g., 1 for Like, -1 for Dislike)
	// example: 1
	ReactionType int `json:"reaction_type"`
	// The ID of the post (if applicable)
	PostID *string `json:"post_id,omitempty"`
	// The ID of the comment (if applicable)
	CommentID *string `json:"comment_id,omitempty"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}