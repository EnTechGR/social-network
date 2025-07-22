package user

import (
	"time"

	"forum/models"
	"forum/repository"
	"forum/utils"
)

func (r *UserRepository) Create(reg models.UserRegistration) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isUsernameTaken(reg.Username); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrUsernameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	_, err = tx.Exec(
		"INSERT INTO user (user_id, username, email, created_at) VALUES (?, ?, ?, ?)",
		userID, reg.Username, reg.Email, createdAt,
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
		ID:        userID,
		Username:  reg.Username,
		Email:     reg.Email,
		CreatedAt: createdAt,
	}, nil
}

func (r *UserRepository) CreateOAuthUser(reg models.UserRegistration, provider, providerUserID, avatarURL, accessToken, refreshToken string, tokenExpiresAt time.Time) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isUsernameTaken(reg.Username); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrUsernameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	_, err = tx.Exec(`INSERT INTO user (user_id, username, email, created_at) VALUES (?, ?, ?, ?)`,
		userID, reg.Username, reg.Email, createdAt,
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
		reg.Username, reg.Email,
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

	return &models.User{
		ID:        userID,
		Username:  reg.Username,
		Email:     reg.Email,
		CreatedAt: createdAt,
	}, nil
}

func (r *UserRepository) isEmailTaken(email string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) isUsernameTaken(username string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", username).Scan(&count)
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

