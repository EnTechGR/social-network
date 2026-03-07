package repository

import (
	"database/sql"
	"time"

	"social-network/pkg/models"
	"social-network/pkg/utils"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// GetByID retrieves a post by ID
func (r *PostRepository) GetByID(postID string) (*models.Post, error) {
	row := r.db.QueryRow(`SELECT post_id, user_id, group_id, visibility, title, content, created_at, updated_at FROM posts WHERE post_id = ?`, postID)
	var p models.Post
	if err := row.Scan(&p.ID, &p.UserID, &p.GroupID, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) GetAllPosts() ([]models.Post, error) {
	rows, err := r.db.Query(`
		SELECT post_id, user_id, group_id, visibility, title, content, created_at, updated_at
		FROM posts 
		WHERE group_id IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.GroupID, &post.Visibility, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Create inserts a new post into the database
// UPDATED: Now handles visibility field and group_id
func (r *PostRepository) Create(post models.Post) (*models.Post, error) {
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()

	// Default visibility to 'public' if not set
	if post.Visibility == "" {
		post.Visibility = "public"
	}

	// Insert post with visibility and group_id
	_, err := r.db.Exec(
		`INSERT INTO posts (post_id, user_id, group_id, visibility, title, content, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		post.ID, post.UserID, post.GroupID, post.Visibility, post.Title, post.Content, post.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetPostsByUser returns all posts by a specific user (excluding group posts)
func (r *PostRepository) GetPostsByUser(userID string) ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.nickname, p.group_id, p.visibility, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		WHERE p.user_id = ? AND p.group_id IS NULL
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
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

// GetPostsWithUsers returns all posts with user information (excluding group posts)
func (r *PostRepository) GetPostsWithUsers() ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.nickname, p.group_id, p.visibility, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		WHERE p.group_id IS NULL
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.group_id, p.title, p.content, p.created_at, p.updated_at
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.group_id, p.visibility, p.title, p.content, p.created_at, p.updated_at
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetPostsCommentedByUser returns posts where the user has left comments
func (r *PostRepository) GetPostsCommentedByUser(userID string) ([]models.PostWithUser, error) {
	query := `
		SELECT DISTINCT p.post_id, p.user_id, u.nickname, p.group_id, p.visibility, p.title, p.content, p.created_at, p.updated_at
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// UpdateVisibility updates the visibility of a post
func (r *PostRepository) UpdateVisibility(postID, visibility string) error {
	result, err := r.db.Exec(
		`UPDATE posts SET visibility = ?, updated_at = ? WHERE post_id = ?`,
		visibility, time.Now(), postID,
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

// AddAllowedUser adds a user to the allowed viewers list of a post
func (r *PostRepository) AddAllowedUser(postID, userID string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO post_allowed_users (post_id, user_id) VALUES (?, ?)`,
		postID, userID,
	)
	return err
}

// RemoveAllowedUser removes a user from the allowed viewers list of a post
func (r *PostRepository) RemoveAllowedUser(postID, userID string) error {
	_, err := r.db.Exec(
		`DELETE FROM post_allowed_users WHERE post_id = ? AND user_id = ?`,
		postID, userID,
	)
	return err
}

// GetAllowedUsers returns all users allowed to view a post
func (r *PostRepository) GetAllowedUsers(postID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT user_id FROM post_allowed_users WHERE post_id = ?`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// CheckIfFollower checks if followerID follows targetUserID
func (r *PostRepository) CheckIfFollower(followerID, targetUserID string) (bool, error) {
	row := r.db.QueryRow(`
		SELECT COUNT(*) FROM follow_relationships 
		WHERE follower_id = ? AND followee_id = ? AND status = 'accepted'
	`, followerID, targetUserID)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// ========== GROUP POST METHODS ==========

// CheckGroupMembership verifies if a user is an active member of a group
func (r *PostRepository) CheckGroupMembership(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM group_members 
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetGroupPosts returns all posts for a specific group
// Only returns posts if the requesting user is a member of the group
func (r *PostRepository) GetGroupPosts(groupID, requesterUserID string) ([]models.PostWithUser, error) {
	// First check if requester is a member
	isMember, err := r.CheckGroupMembership(groupID, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return []models.PostWithUser{}, nil // Return empty slice for non-members
	}

	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.nickname, p.group_id, g.title as group_name, 
		       p.visibility, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN groups g ON p.group_id = g.group_id
		WHERE p.group_id = ?
		ORDER BY p.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var p models.PostWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.GroupName, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetGroupPostByID retrieves a specific group post
// Returns nil if the post doesn't exist or if the requester is not a group member
func (r *PostRepository) GetGroupPostByID(postID, requesterUserID string) (*models.PostWithUser, error) {
	row := r.db.QueryRow(`
		SELECT p.post_id, p.user_id, u.nickname, p.group_id, g.title as group_name,
		       p.visibility, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN groups g ON p.group_id = g.group_id
		WHERE p.post_id = ? AND p.group_id IS NOT NULL
	`, postID)

	var p models.PostWithUser
	if err := row.Scan(&p.ID, &p.UserID, &p.Nickname, &p.GroupID, &p.GroupName, &p.Visibility, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Check if requester is a member of the group
	if p.GroupID != nil {
		isMember, err := r.CheckGroupMembership(*p.GroupID, requesterUserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, nil // Not a member, return nil
		}
	}

	return &p, nil
}

// CreateGroupPost creates a new post within a group
// Only allows creation if the user is an active member of the group
func (r *PostRepository) CreateGroupPost(post models.Post, creatorUserID string) (*models.Post, error) {
	if post.GroupID == nil {
		return nil, sql.ErrNoRows // Group ID is required
	}

	// Verify user is a member of the group
	isMember, err := r.CheckGroupMembership(*post.GroupID, creatorUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, sql.ErrNoRows // Not a member, cannot create post
	}

	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()
	post.UserID = creatorUserID

	// Group posts default to public visibility within the group
	if post.Visibility == "" {
		post.Visibility = "public"
	}

	_, err = r.db.Exec(
		`INSERT INTO posts (post_id, user_id, group_id, visibility, title, content, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		post.ID, post.UserID, post.GroupID, post.Visibility, post.Title, post.Content, post.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}