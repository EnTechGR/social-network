package user

import (
	"database/sql"
	"forum/models"
	"forum/repository"
)

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime
	// ✅ ADDED: first_name, last_name, age, gender to SELECT
	err := r.DB.QueryRow(
		"SELECT user_id, username, email, first_name, last_name, age, gender, created_at FROM user WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime
	// ✅ ADDED: first_name, last_name, age, gender to SELECT
	err := r.DB.QueryRow(
		"SELECT user_id, username, email, first_name, last_name, age, gender, created_at FROM user WHERE user_id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime
	// ✅ ADDED: first_name, last_name, age, gender to SELECT
	err := r.DB.QueryRow(
		"SELECT user_id, username, email, first_name, last_name, age, gender, created_at FROM user WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// ✅ NEW METHOD: GetByEmailOrUsername retrieves a user by either email or username
func (r *UserRepository) GetByEmailOrUsername(login string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		"SELECT user_id, username, email, first_name, last_name, age, gender, created_at FROM user WHERE email = ? OR username = ?",
		login, login,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Age, &user.Gender, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetAuthByUserID retrieves authentication data for a user
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