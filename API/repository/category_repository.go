// repository/category_repository.go
package repository

import (
	"database/sql"
	"forum/models"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	rows, err := r.db.Query("SELECT category_id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// repository/post_repository.go
// func (r *PostRepository) GetPostsByCategoryWithUser(categoryID int) ([]models.PostWithUser, error) {
// 	query := `SELECT p.post_id, p.user_id, u.username, p.category_id, p.title, p.content, p.created_at
// 			  FROM posts p JOIN user u ON p.user_id = u.user_id
// 			  WHERE p.category_id = ?`

// 	rows, err := r.db.Query(query, categoryID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var posts []models.PostWithUser
// 	for rows.Next() {
// 		var p models.PostWithUser
// 		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.CategoryID, &p.Title, &p.Content, &p.CreatedAt); err != nil {
// 			return nil, err
// 		}
// 		posts = append(posts, p)
// 	}
// 	return posts, nil
// }

func (r *CategoryRepository) GetCategoryByID(id int) (*models.Category, error) {
	query := "SELECT category_id, name FROM categories WHERE category_id = ?"
	row := r.db.QueryRow(query, id)
	var category models.Category
	if err := row.Scan(&category.ID, &category.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

	

// // repository/comment_repository.go
func (r *CommentRepository) GetCommentsByPostWithUser(postID string) ([]models.CommentWithUser, error) {
	query := `SELECT c.comment_id, c.post_id, c.user_id, u.username, c.content, c.created_at
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
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *PostRepository) GetPostsByCategoryWithUser(categoryID int) ([]models.PostWithUser, error) {
	rows, err := r.db.Query(`
		SELECT p.post_id, p.user_id, u.username, pc.category_id, p.title, p.content, p.created_at
		FROM posts p
		JOIN post_categories pc ON p.post_id = pc.post_id
		JOIN user u ON p.user_id = u.user_id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithUser
	for rows.Next() {
		var post models.PostWithUser
		err := rows.Scan(&post.ID, &post.UserID, &post.Username, &post.CategoryID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
