package user

import (
	"time"

	"social-network/models"
	"social-network/repository"
	"social-network/utils"
)

// Create creates a new user with all required fields
// Note: Avatar is now handled separately via ImageRepository.UploadUserAvatar()
func (r *UserRepository) Create(reg models.UserRegistration) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isNicknameTaken(reg.Nickname); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrNicknameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	// ✅ UPDATED: Removed avatar_path and avatar_thumbnail_path
	// Avatars are now stored in images_core and linked via user_avatars
	_, err = tx.Exec(
		`INSERT INTO user (
			user_id, nickname, email, first_name, last_name, 
			date_of_birth, about_me, gender, is_private, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, reg.Nickname, reg.Email, reg.FirstName, reg.LastName,
		reg.DateOfBirth, reg.AboutMe, reg.Gender, reg.IsPrivate, createdAt,
	)
	if err != nil {
		return nil, err
	}

	passwordHash, err := utils.HashPassword(reg.Password)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		"INSERT INTO user_auth (user_id, password_hash) VALUES (?, ?)",
		userID, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.User{
		ID:          userID,
		Nickname:    reg.Nickname,
		Email:       reg.Email,
		FirstName:   reg.FirstName,
		LastName:    reg.LastName,
		DateOfBirth: reg.DateOfBirth,
		AboutMe:     reg.AboutMe,
		Gender:      reg.Gender,
		IsPrivate:   reg.IsPrivate,
		CreatedAt:   createdAt,
	}, nil
}

// CreateOAuthUser creates a new user via OAuth with all required fields
// Note: OAuth avatar URLs are handled via ImageRepository
func (r *UserRepository) CreateOAuthUser(reg models.UserRegistration, provider, providerUserID, avatarURL, accessToken, refreshToken string, tokenExpiresAt time.Time) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isNicknameTaken(reg.Nickname); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrNicknameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	// ✅ UPDATED: Removed avatar_url field - avatars now managed via images_core
	_, err = tx.Exec(
		`INSERT INTO user (
			user_id, nickname, email, first_name, last_name, 
			date_of_birth, about_me, gender, is_private, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, reg.Nickname, reg.Email, reg.FirstName, reg.LastName,
		reg.DateOfBirth, reg.AboutMe, reg.Gender, reg.IsPrivate, createdAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO oauth_accounts (
			oauth_id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar_url,
			access_token, refresh_token, token_expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		utils.GenerateUUID(),
		userID, provider, providerUserID,
		reg.Nickname, reg.Email,
		avatarURL,
		accessToken, refreshToken, tokenExpiresAt.Format(time.RFC3339),
		createdAt, createdAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Note: OAuth avatar should be downloaded and uploaded via ImageRepository.UploadUserAvatar()
	// after user creation if avatarURL is provided

	return &models.User{
		ID:          userID,
		Nickname:    reg.Nickname,
		Email:       reg.Email,
		FirstName:   reg.FirstName,
		LastName:    reg.LastName,
		DateOfBirth: reg.DateOfBirth,
		AboutMe:     reg.AboutMe,
		Gender:      reg.Gender,
		IsPrivate:   reg.IsPrivate,
		CreatedAt:   createdAt,
	}, nil
}

func (r *UserRepository) isEmailTaken(email string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) isNicknameTaken(nickname string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE nickname = ?", nickname).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) IsProviderLinked(userID, provider string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM oauth_accounts
		WHERE user_id = ? AND provider = ?
	`, userID, provider).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) LinkOAuthProvider(userID, provider, providerUserID, accessToken, refreshToken string, expiresAt time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO oauth_accounts (
			oauth_id, user_id, provider, provider_user_id,
			access_token, refresh_token, token_expires_at,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(provider, provider_user_id)
		DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_expires_at = excluded.token_expires_at,
			updated_at = excluded.updated_at
	`,
		utils.GenerateUUID(), userID, provider, providerUserID,
		accessToken, refreshToken, expiresAt.Format(time.RFC3339),
		time.Now(), time.Now(),
	)
	return err
}

// DeleteByID deletes a user by ID (used for cleanup on registration failure)
func (r *UserRepository) DeleteByID(userID string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete user_auth first due to foreign key
	_, err = tx.Exec("DELETE FROM user_auth WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Delete user
	_, err = tx.Exec("DELETE FROM user WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}