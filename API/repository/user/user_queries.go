package user

import (
	"database/sql"
	"forum/models"
	"forum/repository"
)

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		"SELECT user_id, username, email, created_at FROM user WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		"SELECT user_id, username, email, created_at FROM user WHERE user_id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		"SELECT user_id, username, email, created_at FROM user WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

func (r *UserRepository) GetAuthByUserID(userID string) (*models.UserAuth, error) {
	var auth models.UserAuth

	err := r.DB.QueryRow(
		"SELECT user_id, password_hash FROM user_auth WHERE user_id = ?",
		userID,
	).Scan(&auth.UserID, &auth.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &auth, nil
}
