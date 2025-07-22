package session

import (
	"database/sql"
	"log"
	"time"

	"forum/models"
	"forum/repository"
	"forum/utils"
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	DB *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

// Create creates a new session for a user
func (r *SessionRepository) Create(userID, ipAddress, csrfToken string) (*models.Session, error) {
	// Generate a new session ID
	sessionID := utils.GenerateSessionToken()
	createdAt := time.Now().UTC()
	expiresAt := utils.CalculateSessionExpiry()

	// Insert or replace the session atomically
	_, err := r.DB.Exec(`INSERT INTO sessions (user_id, session_id, ip_address, created_at, expires_at, csrf_token)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id) DO UPDATE SET
				session_id = excluded.session_id,
				ip_address = excluded.ip_address,
				created_at = excluded.created_at,
				expires_at = excluded.expires_at,
				csrf_token = excluded.csrf_token`,
		userID, sessionID, ipAddress, createdAt.Format(time.RFC3339), expiresAt.Format(time.RFC3339), csrfToken)
	if err != nil {
		return nil, err
	}

	// Return the session object including CSRF token
	session := &models.Session{
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		CSRFToken: csrfToken,
	}

	return session, nil
}

// GetBySessionID retrieves a session by its ID
func (r *SessionRepository) GetBySessionID(sessionID string) (*models.Session, error) {
	log.Printf("GetBySessionID called with sessionID: %s", sessionID)
	var session models.Session
	var createdStr, expiresStr string

	err := r.DB.QueryRow(
		"SELECT user_id, session_id, ip_address, created_at, expires_at, csrf_token FROM sessions WHERE session_id = ?",
		sessionID,
	).Scan(
		&session.UserID,
		&session.SessionID,
		&session.IPAddress,
		&createdStr,
		&expiresStr,
		&session.CSRFToken,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrSessionNotFound
		}
		return nil, err
	}

	session.CreatedAt, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}

	session.ExpiresAt, err = time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		_, _ = r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
		return nil, repository.ErrSessionExpired
	}

	return &session, nil
}

// Delete removes a session
func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}

// Add these methods to your SessionRepository struct

// UpdateLastAccessed updates the last accessed time for a session
func (r *SessionRepository) UpdateLastAccessed(sessionID string) error {
	// Note: The table name is 'sessions', not 'session'.
	_, err := r.DB.Exec(
		"UPDATE sessions SET last_accessed = ? WHERE session_id = ?",
		time.Now(), sessionID,
	)
	return err
}

// DeleteBySessionID deletes a session by session ID
func (r *SessionRepository) DeleteBySessionID(sessionID string) error {
	// Note: The table name is 'sessions', not 'session'.
	_, err := r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}

// DeleteAllUserSessions deletes all sessions for a specific user (useful for logout all devices)
func (r *SessionRepository) DeleteAllUserSessions(userID string) error {
	// Note: The table name is 'sessions', not 'session'.
	_, err := r.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}
