package group

import (
	"database/sql"
	"social-network/models"
	"social-network/utils"
	"time"
)

type GroupEventRepository struct {
	db *sql.DB
}

func NewGroupEventRepository(db *sql.DB) *GroupEventRepository {
	return &GroupEventRepository{db: db}
}

// CreateEvent creates a new event and automatically creates default RSVP options
func (r *GroupEventRepository) CreateEvent(event models.GroupEvent) (*models.GroupEvent, error) {
	event.ID = utils.GenerateUUID()
	event.CreatedAt = time.Now()

	// Start transaction to create event and options atomically
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert event
	_, err = tx.Exec(
		`INSERT INTO group_events (event_id, group_id, creator_id, title, description, event_time, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.GroupID, event.CreatorID, event.Title, event.Description, event.EventTime, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Create default RSVP options
	options := []string{"going", "not going", "maybe"}
	for _, label := range options {
		optionID := utils.GenerateUUID()
		_, err = tx.Exec(
			`INSERT INTO group_event_options (option_id, event_id, label) VALUES (?, ?, ?)`,
			optionID, event.ID, label,
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEventByID retrieves an event by ID
func (r *GroupEventRepository) GetEventByID(eventID string) (*models.GroupEvent, error) {
	var e models.GroupEvent
	err := r.db.QueryRow(
		`SELECT event_id, group_id, creator_id, title, description, event_time, created_at 
		FROM group_events WHERE event_id = ?`,
		eventID,
	).Scan(&e.ID, &e.GroupID, &e.CreatorID, &e.Title, &e.Description, &e.EventTime, &e.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

// GetEventWithDetails retrieves an event with all RSVP details
func (r *GroupEventRepository) GetEventWithDetails(eventID, userID string) (*models.GroupEventWithDetails, error) {
	var e models.GroupEventWithDetails

	// Get event details
	err := r.db.QueryRow(
		`SELECT 
			ge.event_id, ge.group_id, g.title, ge.creator_id, u.nickname,
			ge.title, ge.description, ge.event_time, ge.created_at
		FROM group_events ge
		INNER JOIN groups g ON ge.group_id = g.group_id
		INNER JOIN user u ON ge.creator_id = u.user_id
		WHERE ge.event_id = ?`,
		eventID,
	).Scan(&e.ID, &e.GroupID, &e.GroupTitle, &e.CreatorID, &e.CreatorNickname, &e.Title, &e.Description, &e.EventTime, &e.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get options with vote counts
	rows, err := r.db.Query(
		`SELECT 
			geo.option_id, geo.label, 
			COUNT(gev.user_id) as vote_count,
			SUM(CASE WHEN gev.user_id = ? THEN 1 ELSE 0 END) as user_voted
		FROM group_event_options geo
		LEFT JOIN group_event_votes gev ON geo.option_id = gev.option_id
		WHERE geo.event_id = ?
		GROUP BY geo.option_id
		ORDER BY 
			CASE geo.label 
				WHEN 'going' THEN 1 
				WHEN 'not going' THEN 2 
				WHEN 'maybe' THEN 3 
			END`,
		userID, eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	e.Options = []models.EventOptionWithVotes{}
	var userVoteOptionID *string
	for rows.Next() {
		var opt models.EventOptionWithVotes
		var userVoted int
		if err := rows.Scan(&opt.ID, &opt.Label, &opt.VoteCount, &userVoted); err != nil {
			return nil, err
		}
		opt.EventID = eventID
		opt.UserVoted = userVoted > 0
		if opt.UserVoted {
			userVoteOptionID = &opt.ID
		}
		e.Options = append(e.Options, opt)
	}
	e.UserVote = userVoteOptionID

	return &e, rows.Err()
}

// GetGroupEvents retrieves all events for a group
func (r *GroupEventRepository) GetGroupEvents(groupID string) ([]models.GroupEvent, error) {
	rows, err := r.db.Query(
		`SELECT event_id, group_id, creator_id, title, description, event_time, created_at 
		FROM group_events 
		WHERE group_id = ?
		ORDER BY event_time ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.GroupEvent
	for rows.Next() {
		var e models.GroupEvent
		if err := rows.Scan(&e.ID, &e.GroupID, &e.CreatorID, &e.Title, &e.Description, &e.EventTime, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetEventOptions retrieves all options for an event
func (r *GroupEventRepository) GetEventOptions(eventID string) ([]models.GroupEventOption, error) {
	rows, err := r.db.Query(
		`SELECT option_id, event_id, label 
		FROM group_event_options 
		WHERE event_id = ?
		ORDER BY 
			CASE label 
				WHEN 'going' THEN 1 
				WHEN 'not going' THEN 2 
				WHEN 'maybe' THEN 3 
			END`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.GroupEventOption
	for rows.Next() {
		var opt models.GroupEventOption
		if err := rows.Scan(&opt.ID, &opt.EventID, &opt.Label); err != nil {
			return nil, err
		}
		options = append(options, opt)
	}
	return options, rows.Err()
}

// VoteOnEvent records or updates a user's vote for an event
func (r *GroupEventRepository) VoteOnEvent(eventID, userID, optionID string) error {
	// Delete any existing vote for this event (user can only have one vote per event)
	_, err := r.db.Exec(
		`DELETE FROM group_event_votes WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	)
	if err != nil {
		return err
	}

	// Insert new vote
	_, err = r.db.Exec(
		`INSERT INTO group_event_votes (event_id, user_id, option_id, created_at) 
		VALUES (?, ?, ?, ?)`,
		eventID, userID, optionID, time.Now(),
	)
	return err
}

// RemoveVote removes a user's vote from an event
func (r *GroupEventRepository) RemoveVote(eventID, userID string) error {
	result, err := r.db.Exec(
		`DELETE FROM group_event_votes WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
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

// GetUserVote retrieves a user's vote for an event
func (r *GroupEventRepository) GetUserVote(eventID, userID string) (*string, error) {
	var optionID string
	err := r.db.QueryRow(
		`SELECT option_id FROM group_event_votes WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	).Scan(&optionID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &optionID, nil
}

// GetVoteStats returns vote statistics for an event
func (r *GroupEventRepository) GetVoteStats(eventID string) ([]models.EventOptionWithVotes, error) {
	rows, err := r.db.Query(
		`SELECT 
			geo.option_id, geo.event_id, geo.label, COUNT(gev.user_id) as vote_count
		FROM group_event_options geo
		LEFT JOIN group_event_votes gev ON geo.option_id = gev.option_id
		WHERE geo.event_id = ?
		GROUP BY geo.option_id
		ORDER BY 
			CASE geo.label 
				WHEN 'going' THEN 1 
				WHEN 'not going' THEN 2 
				WHEN 'maybe' THEN 3 
			END`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.EventOptionWithVotes
	for rows.Next() {
		var stat models.EventOptionWithVotes
		if err := rows.Scan(&stat.ID, &stat.EventID, &stat.Label, &stat.VoteCount); err != nil {
			return nil, err
		}
		stat.UserVoted = false // Not checking specific user here
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// DeleteEvent deletes an event (CASCADE will handle options and votes)
func (r *GroupEventRepository) DeleteEvent(eventID string) error {
	result, err := r.db.Exec(`DELETE FROM group_events WHERE event_id = ?`, eventID)
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

// IsUserEventCreator checks if a user created the event
func (r *GroupEventRepository) IsUserEventCreator(eventID, userID string) (bool, error) {
	var creatorID string
	err := r.db.QueryRow(
		`SELECT creator_id FROM group_events WHERE event_id = ?`,
		eventID,
	).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return creatorID == userID, nil
}