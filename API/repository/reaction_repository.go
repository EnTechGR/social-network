package repository

import (
	"database/sql"
	"errors"
	"forum/models"
	"time"
)

type ReactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

func (r *ReactionRepository) GetAllReactions() ([]models.Reaction, error) {
	rows, err := r.db.Query(`
		SELECT user_id, reaction_type, comment_id, post_id, created_at 
		FROM reactions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []models.Reaction
	for rows.Next() {
		var react models.Reaction
		err := rows.Scan(&react.UserID, &react.Type, &react.CommentID, &react.PostID, &react.CreatedAt)
		if err != nil {
			return nil, err
		}
		reactions = append(reactions, react)
	}

	return reactions, nil
}

// ToggleReaction adds or updates a reaction. If the same reaction already
// exists for the user and target, it is removed.
func (r *ReactionRepository) ToggleReaction(userID, targetType, targetID string, reactionType int) error {
	switch targetType {
	case "post":
		var existing int
		err := r.db.QueryRow(`SELECT reaction_type FROM reactions WHERE user_id = ? AND post_id = ?`, userID, targetID).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == sql.ErrNoRows {
			_, err = r.db.Exec(`INSERT INTO reactions (user_id, post_id, reaction_type, created_at) VALUES (?, ?, ?, ?)`, userID, targetID, reactionType, time.Now())
			return err
		}
		if existing == reactionType {
			_, err = r.db.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`, userID, targetID)
			return err
		}
		_, err = r.db.Exec(`UPDATE reactions SET reaction_type = ?, created_at = ? WHERE user_id = ? AND post_id = ?`, reactionType, time.Now(), userID, targetID)
		return err
	case "comment":
		var existing int
		err := r.db.QueryRow(`SELECT reaction_type FROM reactions WHERE user_id = ? AND comment_id = ?`, userID, targetID).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == sql.ErrNoRows {
			_, err = r.db.Exec(`INSERT INTO reactions (user_id, comment_id, reaction_type, created_at) VALUES (?, ?, ?, ?)`, userID, targetID, reactionType, time.Now())
			return err
		}
		if existing == reactionType {
			_, err = r.db.Exec(`DELETE FROM reactions WHERE user_id = ? AND comment_id = ?`, userID, targetID)
			return err
		}
		_, err = r.db.Exec(`UPDATE reactions SET reaction_type = ?, created_at = ? WHERE user_id = ? AND comment_id = ?`, reactionType, time.Now(), userID, targetID)
		return err
	}
	return errors.New("invalid target type")
}

// GetReactionsByPostWithUser returns reactions for a post with usernames
func (r *ReactionRepository) GetReactionsByPostWithUser(postID string) ([]models.ReactionWithUser, error) {
	query := `SELECT r.user_id, u.username, r.reaction_type, r.post_id, r.created_at
                          FROM reactions r JOIN user u ON r.user_id = u.user_id
                          WHERE r.post_id = ?`
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []models.ReactionWithUser
	for rows.Next() {
		var rr models.ReactionWithUser
		if err := rows.Scan(&rr.UserID, &rr.Username, &rr.ReactionType, &rr.PostID, &rr.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, rr)
	}
	return reactions, nil
}

// GetReactionsByCommentWithUser returns reactions for a comment with usernames
func (r *ReactionRepository) GetReactionsByCommentWithUser(commentID string) ([]models.ReactionWithUser, error) {
	query := `SELECT r.user_id, u.username, r.reaction_type, r.comment_id, r.created_at
                          FROM reactions r JOIN user u ON r.user_id = u.user_id
                          WHERE r.comment_id = ?`
	rows, err := r.db.Query(query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []models.ReactionWithUser
	for rows.Next() {
		var rr models.ReactionWithUser
		if err := rows.Scan(&rr.UserID, &rr.Username, &rr.ReactionType, &rr.CommentID, &rr.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, rr)
	}
	return reactions, nil
}
