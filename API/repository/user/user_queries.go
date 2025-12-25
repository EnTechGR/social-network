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

	// UPDATED: Replaced username/age with nickname/date_of_birth and added new profile fields
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
                avatar_url, about_me, gender, is_private, created_at 
         FROM user WHERE email = ?`,
		email,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName, 
		&user.DateOfBirth, &user.AvatarURL, &user.AboutMe, &user.Gender, 
		&user.IsPrivate, &createdAt,
	)

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

	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
                avatar_url, about_me, gender, is_private, created_at 
         FROM user WHERE user_id = ?`,
		id,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName, 
		&user.DateOfBirth, &user.AvatarURL, &user.AboutMe, &user.Gender, 
		&user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByNickname retrieves a user by nickname (Renamed from GetByUsername)
func (r *UserRepository) GetByNickname(nickname string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
                avatar_url, about_me, gender, is_private, created_at 
         FROM user WHERE nickname = ?`,
		nickname,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName, 
		&user.DateOfBirth, &user.AvatarURL, &user.AboutMe, &user.Gender, 
		&user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByEmailOrNickname retrieves a user by either email or nickname (Updated name)
func (r *UserRepository) GetByEmailOrNickname(login string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
                avatar_url, about_me, gender, is_private, created_at 
         FROM user WHERE email = ? OR nickname = ?`,
		login, login,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName, 
		&user.DateOfBirth, &user.AvatarURL, &user.AboutMe, &user.Gender, 
		&user.IsPrivate, &createdAt,
	)

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