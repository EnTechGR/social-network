package oauth

import (
	"database/sql"
	"forum/models"
	"forum/repository"
	"time"
)

// NewOAuthRepository creates a new OAuthRepository
func NewOAuthRepository(db *sql.DB) *OAuthRepository {
	return &OAuthRepository{DB: db}
}

// OAuthRepository handles OAuth-related database operations
type OAuthRepository struct {
	DB *sql.DB
}

// CreateOAuthAccount creates a new OAuth account
func (r *OAuthRepository) CreateOAuthAccount(account *models.OAuthAccount) error {
	now := time.Now().UTC()
	account.CreatedAt = now
	account.UpdatedAt = now

	// Corrected column names: provider_email, provider_username, token_expires_at
	_, err := r.DB.Exec(`
		INSERT INTO oauth_accounts (oauth_id, user_id, provider, provider_user_id, provider_email, provider_username, provider_avatar_url, access_token, refresh_token, token_expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		account.ID, account.UserID, account.Provider, account.ProviderUserID, account.Email, account.Name, account.AvatarURL, // Assuming account.Email is provider_email, account.Name is provider_username
		account.AccessToken, account.RefreshToken, account.TokenExpiry.Format(time.RFC3339),
		account.CreatedAt.Format(time.RFC3339), account.UpdatedAt.Format(time.RFC3339))

	return err
}

// GetOAuthAccountByProvider retrieves an OAuth account by provider and provider user ID
func (r *OAuthRepository) GetOAuthAccountByProvider(provider, providerUserID string) (*models.OAuthAccount, error) {
	var account models.OAuthAccount
	var createdStr, updatedStr, expiryStr string

	// Corrected column names in SELECT statement: oauth_id, provider_email, provider_username, token_expires_at
	err := r.DB.QueryRow(`
		SELECT oauth_id, user_id, provider, provider_user_id, provider_email, provider_username, provider_avatar_url, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts WHERE provider = ? AND provider_user_id = ?`,
		provider, providerUserID,
	).Scan(
		&account.ID, &account.UserID, &account.Provider, &account.ProviderUserID,
		&account.Email, &account.Name, &account.AvatarURL, &account.AccessToken,
		&account.RefreshToken, &expiryStr, &createdStr, &updatedStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrOAuthAccountNotFound
		}
		return nil, err
	}

	// Parse timestamps
	account.CreatedAt, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}

	account.UpdatedAt, err = time.Parse(time.RFC3339, updatedStr)
	if err != nil {
		return nil, err
	}

	account.TokenExpiry, err = time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetOAuthAccountsByUserID retrieves all OAuth accounts for a user
func (r *OAuthRepository) GetOAuthAccountsByUserID(userID string) ([]models.OAuthAccount, error) {
	// Corrected column names in SELECT statement: oauth_id, provider_email, provider_username
	rows, err := r.DB.Query(`
		SELECT oauth_id, user_id, provider, provider_user_id, provider_email, provider_username, provider_avatar_url, created_at, updated_at
		FROM oauth_accounts WHERE user_id = ?`,
		userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.OAuthAccount
	for rows.Next() {
		var account models.OAuthAccount
		var createdStr, updatedStr string

		err := rows.Scan(
			&account.ID, &account.UserID, &account.Provider, &account.ProviderUserID,
			&account.Email, &account.Name, &account.AvatarURL, &createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		account.CreatedAt, err = time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return nil, err
		}

		account.UpdatedAt, err = time.Parse(time.RFC3339, updatedStr)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// UpdateOAuthAccount updates an existing OAuth account
func (r *OAuthRepository) UpdateOAuthAccount(account *models.OAuthAccount) error {
	account.UpdatedAt = time.Now().UTC()

	// Corrected column names in UPDATE statement: provider_email, provider_username, token_expires_at
	_, err := r.DB.Exec(`
		UPDATE oauth_accounts 
		SET provider_email = ?, provider_username = ?, provider_avatar_url = ?, access_token = ?, refresh_token = ?, token_expires_at = ?, updated_at = ?
		WHERE oauth_id = ?`, // Use oauth_id for WHERE clause
		account.Email, account.Name, account.AvatarURL, account.AccessToken, // Assuming account.Email is provider_email, account.Name is provider_username
		account.RefreshToken, account.TokenExpiry.Format(time.RFC3339),
		account.UpdatedAt.Format(time.RFC3339), account.ID)

	return err
}

// DeleteOAuthAccount removes an OAuth account
func (r *OAuthRepository) DeleteOAuthAccount(userID, provider string) error {
	result, err := r.DB.Exec("DELETE FROM oauth_accounts WHERE user_id = ? AND provider = ?", userID, provider)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrOAuthAccountNotFound
	}

	return nil
}

// CheckOAuthAccountExists checks if an OAuth account exists for a provider and user
func (r *OAuthRepository) CheckOAuthAccountExists(provider, providerUserID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(
		"SELECT COUNT(*) FROM oauth_accounts WHERE provider = ? AND provider_user_id = ?",
		provider, providerUserID,
	).Scan(&count)

	return count > 0, err
}

// LinkOAuthAccount links an OAuth account to an existing user
func (r *OAuthRepository) LinkOAuthAccount(userID string, account *models.OAuthAccount) error {
	// Check if this OAuth account is already linked
	userIDLinked, err := r.GetUserByOAuthAccount(account.Provider, account.ProviderUserID)
	if err != nil && err != repository.ErrOAuthAccountNotFound {
		return err
	}

	if userIDLinked != "" && userIDLinked != userID {
		// Linked to another user
		return repository.ErrOAuthAccountExists
	}

	// Set user ID and create the account
	account.UserID = userID
	return r.CreateOAuthAccount(account)
}

// GetUserByOAuthAccount retrieves a user by their OAuth account
func (r *OAuthRepository) GetUserByOAuthAccount(provider, providerUserID string) (string, error) {
	var userID string
	err := r.DB.QueryRow(
		"SELECT user_id FROM oauth_accounts WHERE provider = ? AND provider_user_id = ?",
		provider, providerUserID,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", repository.ErrOAuthAccountNotFound
		}
		return "", err
	}

	return userID, nil
}

// CleanupExpiredOAuthStates removes expired OAuth states
func (r *OAuthRepository) CleanupExpiredOAuthStates() error {
	_, err := r.DB.Exec("DELETE FROM oauth_states WHERE expires_at < ?", time.Now().Format(time.RFC3339))
	return err
}
