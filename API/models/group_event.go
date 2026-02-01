package models

import "time"

// GroupEvent represents an event created within a group
// swagger:model GroupEvent
type GroupEvent struct {
	// The unique identifier for the event (UUID)
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group this event belongs to
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The ID of the user who created the event
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	CreatorID string `json:"creator_id"`
	// The title of the event
	// example: Photo Walk in Central Park
	Title string `json:"title"`
	// The description of the event
	// example: Join us for a scenic photo walk through Central Park. Bring your camera!
	Description *string `json:"description,omitempty"`
	// The date and time of the event
	// example: 2025-12-30T14:00:00Z
	EventTime time.Time `json:"event_time"`
	// The timestamp when the event was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// GroupEventOption represents an RSVP option for an event (going, not going, maybe)
// swagger:model GroupEventOption
type GroupEventOption struct {
	// The unique identifier for the option (UUID)
	// example: o1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the event this option belongs to
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	EventID string `json:"event_id"`
	// The label for this option: going, not going, maybe
	// example: going
	Label string `json:"label"`
}

// GroupEventVote represents a user's RSVP vote for an event
// swagger:model GroupEventVote
type GroupEventVote struct {
	// The ID of the event
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	EventID string `json:"event_id"`
	// The ID of the user who voted
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The ID of the option the user chose
	// example: o1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	OptionID string `json:"option_id"`
	// The timestamp when the vote was cast
	// example: 2025-12-29T19:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// EventOptionWithVotes represents an option with vote count
// swagger:model EventOptionWithVotes
type EventOptionWithVotes struct {
	// The unique identifier for the option (UUID)
	// example: o1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the event this option belongs to
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	EventID string `json:"event_id"`
	// The label for this option: going, not going, maybe
	// example: going
	Label string `json:"label"`
	// The number of votes for this option
	// example: 15
	VoteCount int `json:"vote_count"`
	// Whether the current user voted for this option
	// example: true
	UserVoted bool `json:"user_voted"`
}

// GroupEventWithDetails represents an event with all related information
// swagger:model GroupEventWithDetails
type GroupEventWithDetails struct {
	// The unique identifier for the event (UUID)
	// example: e1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	ID string `json:"id"`
	// The ID of the group this event belongs to
	// example: g1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	GroupID string `json:"group_id"`
	// The title of the group
	// example: Photography Enthusiasts
	GroupTitle string `json:"group_title"`
	// The ID of the user who created the event
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	CreatorID string `json:"creator_id"`
	// The display name of the creator
	// example: photo_master
	CreatorNickname string `json:"creator_nickname"`
	// The title of the event
	// example: Photo Walk in Central Park
	Title string `json:"title"`
	// The description of the event
	// example: Join us for a scenic photo walk through Central Park. Bring your camera!
	Description *string `json:"description,omitempty"`
	// The date and time of the event
	// example: 2025-12-30T14:00:00Z
	EventTime time.Time `json:"event_time"`
	// The timestamp when the event was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The RSVP options with vote counts
	Options []EventOptionWithVotes `json:"options"`
	// The option ID the current user voted for (if any)
	// example: o1a2b3c4-d5e6-f7g8-h9i0-j1k2l3m4n5o6
	UserVote *string `json:"user_vote,omitempty"`
}