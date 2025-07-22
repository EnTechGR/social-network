package models

import "time"

// Session represents a user session
type Session struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	CSRFToken string    `json:"csrf_token"` // CSRF token for security
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
