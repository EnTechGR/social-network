package group

import (
	"database/sql"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"time"
)

type GroupEventRepository struct {
	db *sql.DB
}

func NewGroupEventRepository(db *sql.DB) *GroupEventRepository {
	return &GroupEventRepository{db: db}
}

// CreateEvent inserts a new event.
// Options are implicit (going / not going / maybe) — no per-event rows needed.
func (r *GroupEventRepository) CreateEvent(event models.GroupEvent) (*models.GroupEvent, error) {
	event.ID = utils.GenerateUUID()
	event.CreatedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO group_events (event_id, group_id, creator_id, title, description, event_time, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.GroupID, event.CreatorID, event.Title, event.Description, event.EventTime, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEventByID retrieves an event by ID.
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

// GetEventWithDetails retrieves an event together with per-choice vote counts.
// A CTE materialises the three fixed choices so that labels with zero votes
// still appear in the result — no LEFT JOIN to a separate options table required.
func (r *GroupEventRepository) GetEventWithDetails(eventID, userID string) (*models.GroupEventWithDetails, error) {
	var e models.GroupEventWithDetails

	// ── event row ────────────────────────────────────────────────────────────
	err := r.db.QueryRow(
		`SELECT
			ge.event_id, ge.group_id, g.title, ge.creator_id, u.nickname,
			ge.title, ge.description, ge.event_time, ge.created_at
		 FROM   group_events ge
		 INNER JOIN groups g ON ge.group_id  = g.group_id
		 INNER JOIN user  u ON ge.creator_id = u.user_id
		 WHERE  ge.event_id = ?`,
		eventID,
	).Scan(&e.ID, &e.GroupID, &e.GroupTitle, &e.CreatorID, &e.CreatorNickname,
		&e.Title, &e.Description, &e.EventTime, &e.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// ── choices + vote counts ────────────────────────────────────────────────
	// The CTE `all_choices` is the single source of truth for the three labels.
	// The LEFT JOIN pulls in only the rows that exist; COUNT / SUM handle the
	// zero-vote case transparently.
	rows, err := r.db.Query(
		`WITH all_choices(label, sort_order) AS (
			VALUES ('going', 1), ('not going', 2), ('maybe', 3)
		)
		SELECT ac.label,
			   COUNT(gev.user_id)                                    AS vote_count,
			   SUM(CASE WHEN gev.user_id = ? THEN 1 ELSE 0 END)    AS user_voted
		FROM   all_choices ac
		LEFT JOIN group_event_votes gev
			   ON gev.choice   = ac.label
			  AND gev.event_id = ?
		GROUP BY ac.label
		ORDER BY ac.sort_order`,
		userID, eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	e.Options = []models.EventOptionWithVotes{}
	var userVoteLabel *string

	for rows.Next() {
		var opt models.EventOptionWithVotes
		var userVoted int
		if err := rows.Scan(&opt.Label, &opt.VoteCount, &userVoted); err != nil {
			return nil, err
		}
		opt.UserVoted = userVoted > 0
		if opt.UserVoted {
			label := opt.Label // copy — opt is reused
			userVoteLabel = &label
		}
		e.Options = append(e.Options, opt)
	}
	e.UserVote = userVoteLabel

	return &e, rows.Err()
}

// GetGroupEvents retrieves all events for a group ordered by event_time.
func (r *GroupEventRepository) GetGroupEvents(groupID string) ([]models.GroupEvent, error) {
	rows, err := r.db.Query(
		`SELECT event_id, group_id, creator_id, title, description, event_time, created_at
		 FROM   group_events
		 WHERE  group_id = ?
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

// VoteOnEvent records (or replaces) a user's RSVP choice for an event.
// `choice` must be one of: "going", "not going", "maybe".
// The CHECK constraint on the table is the final safety net, but the handler
// validates before we ever reach this point.
func (r *GroupEventRepository) VoteOnEvent(eventID, userID, choice string) error {
	// Remove previous vote if any (one vote per user per event).
	_, err := r.db.Exec(
		`DELETE FROM group_event_votes WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO group_event_votes (event_id, user_id, choice, created_at)
		 VALUES (?, ?, ?, ?)`,
		eventID, userID, choice, time.Now(),
	)
	return err
}

// RemoveVote deletes a user's vote from an event entirely.
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

// GetUserVote returns the choice label the user picked for this event,
// or nil when the user has not yet voted.
func (r *GroupEventRepository) GetUserVote(eventID, userID string) (*string, error) {
	var choice string
	err := r.db.QueryRow(
		`SELECT choice FROM group_event_votes WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	).Scan(&choice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &choice, nil
}

// GetVoteStats returns the vote breakdown for an event without caller context
// (no user_voted flag).  Useful for analytics or admin views.
func (r *GroupEventRepository) GetVoteStats(eventID string) ([]models.EventOptionWithVotes, error) {
	rows, err := r.db.Query(
		`WITH all_choices(label, sort_order) AS (
			VALUES ('going', 1), ('not going', 2), ('maybe', 3)
		)
		SELECT ac.label,
			   COUNT(gev.user_id) AS vote_count
		FROM   all_choices ac
		LEFT JOIN group_event_votes gev
			   ON gev.choice   = ac.label
			  AND gev.event_id = ?
		GROUP BY ac.label
		ORDER BY ac.sort_order`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.EventOptionWithVotes
	for rows.Next() {
		var stat models.EventOptionWithVotes
		if err := rows.Scan(&stat.Label, &stat.VoteCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// DeleteEvent removes an event.  CASCADE on group_event_votes handles cleanup.
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

// IsUserEventCreator returns true when the supplied user is the creator of the event.
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