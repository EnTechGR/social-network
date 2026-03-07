package repository

import (
	"database/sql"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"time"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(commentID string) (*models.Comment, error) {
	row := r.db.QueryRow(`SELECT comment_id, post_id, user_id, content, created_at, updated_at FROM comments WHERE comment_id = ?`, commentID)
	var c models.Comment
	if err := row.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CommentRepository) GetAllComments() ([]models.Comment, error) {
	rows, err := r.db.Query(`
		SELECT comment_id, post_id, user_id, content, created_at, updated_at 
		FROM comments ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// Create inserts a new comment into the database
func (r *CommentRepository) Create(comment models.Comment) (*models.Comment, error) {
	comment.ID = utils.GenerateUUID()
	comment.CreatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO comments (comment_id, post_id, user_id, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		comment.ID, comment.PostID, comment.UserID, comment.Content, comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// // repository/comment_repository.go
func (r *CommentRepository) GetCommentsByPostWithUser(postID string) ([]models.CommentWithUser, error) {
	query := `SELECT c.comment_id, c.post_id, c.user_id, u.nickname, c.content, c.created_at, c.updated_at
			  FROM comments c JOIN user u ON c.user_id = u.user_id
			  WHERE c.post_id = ?`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.CommentWithUser
	for rows.Next() {
		var c models.CommentWithUser
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Nickname, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// GetCommentByID retrieves a comment by its ID
func (r *CommentRepository) GetCommentByID(commentID string) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.QueryRow(`
		SELECT comment_id, post_id, user_id, content, created_at, updated_at 
		FROM comments WHERE comment_id = ?`, commentID).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// UpdateComment updates the content of a comment and sets updated_at
func (r *CommentRepository) UpdateComment(commentID string, content *string) error {
	_, err := r.db.Exec(`UPDATE comments SET content = ?, updated_at = ? WHERE comment_id = ?`, content, time.Now(), commentID)
	return err
}

// SoftDeleteComment sets content to NULL and updates updated_at
func (r *CommentRepository) SoftDeleteComment(commentID string) error {
	_, err := r.db.Exec(`UPDATE comments SET content = NULL, updated_at = ? WHERE comment_id = ?`, time.Now(), commentID)
	return err
}

// CheckCommentAccessForGroupPost verifies if a user can comment on a group post
// Returns true if:
// - The post is not a group post (groupID is NULL), OR
// - The user is an active member of the group
func (r *CommentRepository) CheckCommentAccessForGroupPost(postID, userID string) (bool, error) {
	// First, get the group_id of the post
	var groupID sql.NullString
	err := r.db.QueryRow(`
		SELECT group_id FROM posts WHERE post_id = ?
	`, postID).Scan(&groupID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Post doesn't exist
		}
		return false, err
	}

	// If this is not a group post, allow access
	if !groupID.Valid {
		return true, nil
	}

	// Check if user is a member of the group
	var count int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM group_members 
		WHERE group_id = ? AND user_id = ?
	`, groupID.String, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetCommentsByPostID should be updated to check group membership
// This is a conceptual example - actual implementation depends on your CommentRepository structure
func (r *CommentRepository) GetCommentsByPostIDWithAccess(postID, requesterUserID string) ([]models.Comment, error) {
	// First check if requester has access to view this post's comments
	hasAccess, err := r.CheckCommentAccessForGroupPost(postID, requesterUserID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return []models.Comment{}, nil // Return empty slice if no access
	}

	// Your existing query to get comments
	// ... (rest of your implementation)
	return nil, nil // Placeholder
}
