package session

import "time"

// DeleteExpiredSessions removes all expired sessions (cleanup utility)
func (r *SessionRepository) DeleteExpiredSessions() error {
	// Note: The table name is 'sessions', not 'session'.
	_, err := r.DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}
