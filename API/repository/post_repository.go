package repository

import (
	"database/sql"
	"time"

	"forum/models"
	"forum/utils"
)

type PostRepository struct {
	db *sql.DB
}

// checks if the legacy category_id column exists on the posts table
func (r *PostRepository) hasLegacyCategoryColumn() bool {
	rows, err := r.db.Query(`PRAGMA table_info(posts)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
			if name == "category_id" {
				return true
			}
		}
	}
	return false
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) GetAllPosts() ([]models.Post, error) {
	rows, err := r.db.Query(`
		SELECT post_id, user_id, title, content, created_at, updated_at
                FROM posts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		//err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Create inserts a new post into the database
// func (r *PostRepository) Create(post models.Post) (*models.Post, error) {
func (r *PostRepository) Create(post models.Post, categoryIDs []int) (*models.Post, error) {
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()
	// _, err := r.db.Exec(`INSERT INTO posts (post_id, user_id, category_id, title, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
	// 	post.ID, post.UserID, post.CategoryID, post.Title, post.Content, post.CreatedAt)

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var insertPost string
	var args []interface{}
	if r.hasLegacyCategoryColumn() {
		if len(categoryIDs) == 0 {
			tx.Rollback()
			return nil, sql.ErrNoRows
		}
		insertPost = `INSERT INTO posts (post_id, user_id, category_id, title, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`
		args = []interface{}{post.ID, post.UserID, categoryIDs[0], post.Title, post.Content, post.CreatedAt}
	} else {
		insertPost = `INSERT INTO posts (post_id, user_id, title, content, created_at) VALUES (?, ?, ?, ?, ?)`
		args = []interface{}{post.ID, post.UserID, post.Title, post.Content, post.CreatedAt}
	}

	_, err = tx.Exec(insertPost, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	stmt, err := tx.Prepare(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	for _, cid := range categoryIDs {
		if _, err := stmt.Exec(post.ID, cid); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostRepository) GetPostsByUser(userID string) ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
        SELECT p.post_id, p.user_id, u.username, p.title, p.content, p.created_at
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *PostRepository) GetCategoriesByPostID(postID string) ([]models.Category, error) {
	rows, err := r.db.Query(`
        SELECT c.category_id, c.name
        FROM categories c
        JOIN post_categories pc ON c.category_id = pc.category_id
        WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// GetPostsReactedByUser returns posts that the given user has reacted to either
// directly or via reactions on comments. The posts are ordered by creation time
// descending.
// func (r *PostRepository) GetPostsReactedByUser(userID string) ([]models.PostWithUser, error) {
// 	query := `SELECT DISTINCT p.post_id, p.user_id, u.username, p.title, p.content, p.created_at
//                 FROM posts p
//                 JOIN user u ON p.user_id = u.user_id
//                 WHERE p.post_id IN (
//                         SELECT post_id FROM reactions WHERE user_id = ? AND post_id IS NOT NULL
//                         UNION
//                         SELECT c.post_id FROM reactions r JOIN comments c ON r.comment_id = c.comment_id WHERE r.user_id = ?
//                 )
//                 ORDER BY p.created_at DESC`

// 	rows, err := r.db.Query(query, userID, userID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var posts []models.PostWithUser
// 	for rows.Next() {
// 		var p models.PostWithUser
// 		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt); err != nil {
// 			return nil, err
// 		}
// 		posts = append(posts, p)
// 	}
// 	return posts, nil
// }


func (r *PostRepository) GetPostsReactedByUser(userID string) ([]models.PostWithUser, error) {
	query := `
		SELECT DISTINCT p.post_id, p.user_id, u.username, p.title, p.content, p.created_at
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		WHERE p.post_id IN (
			SELECT post_id FROM reactions
			WHERE user_id = ? AND reaction_type = 1 AND post_id IS NOT NULL
		)
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// func (r *PostRepository) GetPostsReactedByUser(userID string) ([]models.PostWithUser, error) {
// 	query := `
// 		SELECT DISTINCT p.post_id, p.user_id, u.username, p.title, p.content, p.created_at
// 		FROM posts p
// 		JOIN user u ON p.user_id = u.user_id
// 		WHERE p.post_id IN (
// 			SELECT post_id FROM reactions
// 			WHERE user_id = ? AND reaction_type = 1 AND post_id IS NOT NULL
// 			UNION
// 			SELECT c.post_id FROM reactions r
// 			JOIN comments c ON r.comment_id = c.comment_id
// 			WHERE r.user_id = ? AND r.reaction_type = 1
// 		)
// 		ORDER BY p.created_at DESC
// 	`

// 	rows, err := r.db.Query(query, userID, userID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var posts []models.PostWithUser
// 	for rows.Next() {
// 		var p models.PostWithUser
// 		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt); err != nil {
// 			return nil, err
// 		}
// 		posts = append(posts, p)
// 	}
// 	return posts, nil
// }
