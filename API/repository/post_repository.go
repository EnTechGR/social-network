package repository

import (
	"database/sql"
	"time"

	"social-network/models"
	"social-network/utils"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// GetByID retrieves a post by ID
func (r *PostRepository) GetByID(postID string) (*models.Post, error) {
	row := r.db.QueryRow(`SELECT post_id, user_id, title, content, created_at, updated_at FROM posts WHERE post_id = ?`, postID)
	var p models.Post
	if err := row.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) GetAllPosts() ([]models.Post, error) {
	rows, err := r.db.Query(`
		SELECT post_id, user_id, title, content, created_at, updated_at
		FROM posts 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Create inserts a new post into the database
// UPDATED: Removed category logic - posts no longer need categories
func (r *PostRepository) Create(post models.Post) (*models.Post, error) {
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()

	// Simple insert without categories
	_, err := r.db.Exec(
		`INSERT INTO posts (post_id, user_id, title, content, created_at) 
		VALUES (?, ?, ?, ?, ?)`,
		post.ID, post.UserID, post.Title, post.Content, post.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetPostsByUser returns all posts by a specific user
func (r *PostRepository) GetPostsByUser(userID string) ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.nickname, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// UpdateTitle updates the title of a post
func (r *PostRepository) UpdateTitle(postID, title string) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE posts SET title = ?, updated_at = ? WHERE post_id = ?`,
		title, now, postID,
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

// UpdateContent updates the content of a post
func (r *PostRepository) UpdateContent(postID, content string) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE posts SET content = ?, updated_at = ? WHERE post_id = ?`,
		content, now, postID,
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

// DeleteByID deletes a post by ID
func (r *PostRepository) DeleteByID(postID string) error {
	result, err := r.db.Exec(`DELETE FROM posts WHERE post_id = ?`, postID)
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

// GetPostsWithUsers returns all posts with user information
func (r *PostRepository) GetPostsWithUsers() ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.nickname, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// CheckOwnership verifies if a user owns a specific post
func (r *PostRepository) CheckOwnership(postID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM posts WHERE post_id = ? AND user_id = ?`,
		postID, userID,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetPostsReactedByUser returns posts that the given user has reacted to (liked)
// Returns posts where user has reaction_type = 1 (like)
func (r *PostRepository) GetPostsReactedByUser(userID string) ([]models.PostWithUser, error) {
	query := `
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN reactions r ON p.post_id = r.post_id
		WHERE r.user_id = ? AND r.reaction_type = 1
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetPostsDislikedByUser returns posts that the given user has disliked
// Returns posts where user has reaction_type = 2 (dislike)
func (r *PostRepository) GetPostsDislikedByUser(userID string) ([]models.PostWithUser, error) {
	query := `
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN reactions r ON p.post_id = r.post_id
		WHERE r.user_id = ? AND r.reaction_type = 2
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetPostsCommentedByUser returns posts where the user has left comments
func (r *PostRepository) GetPostsCommentedByUser(userID string) ([]models.PostWithUser, error) {
	query := `
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN comments c ON p.post_id = c.post_id
		WHERE c.user_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}
