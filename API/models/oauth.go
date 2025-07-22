package models

import "time"

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID             int       `json:"id"`
	UserID         string    `json:"user_id"`
	Provider       string    `json:"provider"`         // "google", "github", "discord", etc.
	ProviderUserID string    `json:"provider_user_id"` // The user ID from the OAuth provider
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	AccessToken    string    `json:"-"` // Don't expose in JSON
	RefreshToken   string    `json:"-"` // Don't expose in JSON
	TokenExpiry    time.Time `json:"-"` // Don't expose in JSON
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OAuthState represents OAuth state for CSRF protection
type OAuthState struct {
	State     string    `json:"state"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
}


// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}