// Updated DeleteExpiredSessions method for SessionRepository
// Replace in API/repository/session/session_cleanup.go

package session

import (
	"log"
	"time"
)

// DeleteExpiredSessions removes all expired sessions (cleanup utility)
// Checks both idle timeout (expires_at) and absolute timeout (absolute_expires_at)
func (r *SessionRepository) DeleteExpiredSessions() error {
	now := time.Now()

	// Delete sessions that have exceeded EITHER idle timeout OR absolute timeout
	result, err := r.DB.Exec(`
		DELETE FROM sessions 
		WHERE expires_at < ? OR absolute_expires_at < ?
	`, now, now)

	if err != nil {
		return err
	}

	// Log how many sessions were cleaned up
	if rows, err := result.RowsAffected(); err == nil && rows > 0 {
		log.Printf("[SESSION CLEANUP] Removed %d expired sessions", rows)
	}

	return nil
}
