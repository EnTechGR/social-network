package oauth

import (
	"database/sql"
	"errors"
	"forum/repository"
	"time"
)

// CreateOAuthState creates a new OAuth state for CSRF protection
func (r *OAuthRepository) CreateOAuthState(state, provider, ipAddress string, expiresAt time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO oauth_states (state, provider, ip_address, expires_at)
		VALUES (?, ?, ?, ?)`,
		state, provider, ipAddress, expiresAt.Format(time.RFC3339))
	return err
}

// ValidateOAuthState validates and consumes an OAuth state
func (r *OAuthRepository) ValidateOAuthState(state, provider string) error {
	var storedProvider string
	var expiresStr string

	err := r.DB.QueryRow(
		"SELECT provider, expires_at FROM oauth_states WHERE state = ?",
		state,
	).Scan(&storedProvider, &expiresStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return repository.ErrOAuthStateNotFound
		}
		return err
	}

	// Parse expiration time
	expiresAt, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return err
	}

	// Check if state is expired
	if time.Now().After(expiresAt) {
		// Clean up expired state
		_, _ = r.DB.Exec("DELETE FROM oauth_states WHERE state = ?", state)
		return repository.ErrOAuthStateExpired
	}

	// Validate provider matches
	if storedProvider != provider {
		return errors.New("provider mismatch")
	}

	// Consume the state (delete it to prevent reuse)
	_, err = r.DB.Exec("DELETE FROM oauth_states WHERE state = ?", state)
	return err
}
