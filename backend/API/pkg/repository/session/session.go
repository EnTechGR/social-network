package session

import (
	"database/sql"
	"log"
	"time"

	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	DB *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

// Updated Create method for SessionRepository with User-Agent support
// Replace your existing Create method in API/repository/session/session.go

// Create creates a new session for a user with IP address, User-Agent, and timeouts
func (r *SessionRepository) Create(userID, ipAddress, userAgent, csrfToken string) (*models.Session, error) {
	// Generate a new session ID
	sessionID, err := utils.GenerateSessionToken()
	if err != nil {
		return nil, err
	}
	createdAt := time.Now().UTC()
	expiresAt := utils.CalculateSessionExpiry()              // Idle timeout (30 min)
	absoluteExpiresAt := utils.CalculateAbsoluteSessionExpiry() // Absolute timeout (12 hours)

	// Insert or replace the session atomically
	_, err = r.DB.Exec(`INSERT INTO sessions (
			user_id, session_id, ip_address, user_agent, created_at, expires_at, absolute_expires_at, csrf_token
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			session_id = excluded.session_id,
			ip_address = excluded.ip_address,
			user_agent = excluded.user_agent,
			created_at = excluded.created_at,
			expires_at = excluded.expires_at,
			absolute_expires_at = excluded.absolute_expires_at,
			csrf_token = excluded.csrf_token`,
		userID, sessionID, ipAddress, userAgent,
		createdAt.Format(time.RFC3339), 
		expiresAt.Format(time.RFC3339),
		absoluteExpiresAt.Format(time.RFC3339),
		csrfToken)
	if err != nil {
		return nil, err
	}

	// Return the session object including User-Agent
	session := &models.Session{
		UserID:            userID,
		SessionID:         sessionID,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		CreatedAt:         createdAt,
		ExpiresAt:         expiresAt,
		AbsoluteExpiresAt: absoluteExpiresAt,
		CSRFToken:         csrfToken,
	}

	return session, nil
}

// Updated GetBySessionID method for SessionRepository with User-Agent retrieval
// Replace your existing GetBySessionID method in API/repository/session/session.go

// GetBySessionID retrieves a session by its ID and validates both idle and absolute timeouts
func (r *SessionRepository) GetBySessionID(sessionID string) (*models.Session, error) {
	log.Printf("GetBySessionID called with sessionID: %s", sessionID)
	var session models.Session
	var createdStr, expiresStr, absoluteExpiresStr string

	err := r.DB.QueryRow(
		`SELECT user_id, session_id, ip_address, user_agent, created_at, expires_at, absolute_expires_at, csrf_token 
		 FROM sessions WHERE session_id = ?`,
		sessionID,
	).Scan(
		&session.UserID,
		&session.SessionID,
		&session.IPAddress,
		&session.UserAgent,
		&createdStr,
		&expiresStr,
		&absoluteExpiresStr,
		&session.CSRFToken,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrSessionNotFound
		}
		return nil, err
	}

	// Parse timestamps
	session.CreatedAt, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}

	session.ExpiresAt, err = time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return nil, err
	}

	session.AbsoluteExpiresAt, err = time.Parse(time.RFC3339, absoluteExpiresStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// SECURITY: Check absolute timeout FIRST (highest priority)
	// Force re-authentication after absolute timeout regardless of activity
	if now.After(session.AbsoluteExpiresAt) {
		log.Printf("[SECURITY] Session %s exceeded absolute timeout (created: %v, absolute expiry: %v)", 
			sessionID, session.CreatedAt, session.AbsoluteExpiresAt)
		_, _ = r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
		return nil, repository.ErrSessionExpired
	}

	// Check idle timeout (inactivity timeout)
	if now.After(session.ExpiresAt) {
		log.Printf("[SECURITY] Session %s exceeded idle timeout (last activity expiry: %v)", 
			sessionID, session.ExpiresAt)
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

// Add this method to your SessionRepository in API/repository/session/session.go

// RenewSession extends the idle timeout for an active session (sliding window)
// Returns the new expires_at time and whether renewal was performed
// Will NOT extend past the absolute timeout - returns false if renewal would exceed it
func (r *SessionRepository) RenewSession(sessionID string, absoluteExpiresAt time.Time) (time.Time, bool, error) {
	// Calculate the new idle timeout expiry
	newExpiresAt := utils.CalculateSessionExpiry()
	
	// SECURITY CHECK: Never extend past absolute timeout
	// Even if user is very active, they must re-authenticate after absolute timeout
	if newExpiresAt.After(absoluteExpiresAt) {
		// Renewal would extend past absolute timeout - don't renew
		// Session will expire naturally when it hits absolute timeout
		log.Printf("[SESSION RENEWAL] Session %s renewal blocked - would exceed absolute timeout", sessionID)
		return time.Time{}, false, nil
	}
	
	// Safe to renew - update the expires_at in database
	_, err := r.DB.Exec(`
		UPDATE sessions 
		SET expires_at = ? 
		WHERE session_id = ?
	`, newExpiresAt.Format(time.RFC3339), sessionID)
	
	if err != nil {
		return time.Time{}, false, err
	}
	
	return newExpiresAt, true, nil
}

// RotateCSRFToken generates a new CSRF token and updates it in the session
// Returns the new CSRF token
// This should be called after sensitive operations to limit attack window
func (r *SessionRepository) RotateCSRFToken(sessionID string) (string, error) {
	// Generate new CSRF token
	newCSRFToken, err := utils.GenerateCSRFToken()
	if err != nil {
		return "", err
	}
	
	// Update in database
	_, err = r.DB.Exec(`
		UPDATE sessions 
		SET csrf_token = ? 
		WHERE session_id = ?
	`, newCSRFToken, sessionID)
	
	if err != nil {
		return "", err
	}
	
	log.Printf("[CSRF ROTATION] Rotated CSRF token for session %s", sessionID)
	return newCSRFToken, nil
}

// UpdateCSRFToken updates the CSRF token for a session
// Used when you already have a new token generated
func (r *SessionRepository) UpdateCSRFToken(sessionID, newCSRFToken string) error {
	_, err := r.DB.Exec(`
		UPDATE sessions 
		SET csrf_token = ? 
		WHERE session_id = ?
	`, newCSRFToken, sessionID)
	
	if err != nil {
		return err
	}
	
	return nil
}