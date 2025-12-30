package models

import "time"

// Session represents an active user session and security context
// swagger:model Session
type Session struct {
	// The ID of the user this session belongs to
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The unique session identifier (usually stored in a secure cookie)
	// example: sess_8f2d9a1b3c4e
	SessionID string `json:"session_id"`
	// The CSRF token required for state-changing operations
	// example: csrf_9b8a7c6d5e4f
	CSRFToken string `json:"csrf_token"`
	// The IP address from which the session was initiated
	// example: 192.168.1.50
	IPAddress string `json:"ip_address"`
	// The User-Agent string from the browser/client that created the session
	// Used for session hijacking detection
	// example: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0
	UserAgent string `json:"user_agent"`
	// The timestamp when the session was created
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// The timestamp when the session will expire due to inactivity (idle timeout)
	// example: 2025-12-30T18:00:00Z
	ExpiresAt time.Time `json:"expires_at"`
	// The timestamp when the session will absolutely expire regardless of activity (absolute timeout)
	// This enforces re-authentication even for active users after a maximum period
	// example: 2025-12-30T06:00:00Z
	AbsoluteExpiresAt time.Time `json:"absolute_expires_at"`
}