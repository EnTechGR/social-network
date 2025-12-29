package models

import "time"

// OAuthAccount represents an external account linked to a local user
// swagger:model OAuthAccount
type OAuthAccount struct {
	// Internal database ID
	// example: 1
	ID int `json:"id"`
	// The ID of the local user this account is linked to
	// example: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
	UserID string `json:"user_id"`
	// The OAuth provider name
	// example: google
	Provider string `json:"provider"`
	// The unique user ID provided by the external service
	// example: 102938475647382
	ProviderUserID string `json:"provider_user_id"`
	// The email address associated with the OAuth account
	// example: user@gmail.com
	Email string `json:"email"`
	// The full name from the OAuth provider
	// example: John Doe
	Name string `json:"name"`
	// The URL to the user's avatar image
	// example: https://lh3.googleusercontent.com/a/abc123
	AvatarURL string `json:"avatar_url,omitempty"`
	// AccessToken is hidden from JSON and documentation
	AccessToken string `json:"-"`
	// RefreshToken is hidden from JSON and documentation
	RefreshToken string `json:"-"`
	// TokenExpiry is hidden from JSON and documentation
	TokenExpiry time.Time `json:"-"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: 2025-12-29T18:00:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// OAuthState represents internal state used to prevent CSRF during OAuth flows
// swagger:model OAuthState
type OAuthState struct {
	// The random state string sent to the provider
	// example: 4fb9-a2c3-987d
	State string `json:"state"`
	// The provider intended for this state
	// example: github
	Provider string `json:"provider"`
	// example: 2025-12-29T18:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: 2025-12-29T18:15:00Z
	ExpiresAt time.Time `json:"expires_at"`
	// The IP address that initiated the OAuth request
	// example: 192.168.1.1
	IPAddress string `json:"ip_address"`
}

// OAuthUserInfo represents the data payload received from an OAuth provider
// swagger:model OAuthUserInfo
type OAuthUserInfo struct {
	// External Provider's User ID
	// example: 987654321
	ID string `json:"id"`
	// example: dev_user@example.com
	Email string `json:"email"`
	// example: Developer User
	Name string `json:"name"`
	// example: dev_pro
	Nickname string `json:"nickname,omitempty"`
	// example: https://avatars.githubusercontent.com/u/123
	AvatarURL string `json:"avatar_url,omitempty"`
}

// OAuthLoginResponse provides the URL to redirect the user to for authentication
// swagger:model OAuthLoginResponse
type OAuthLoginResponse struct {
	// The full URL to the provider's authorization page
	// example: https://accounts.google.com/o/oauth2/auth?client_id=...
	AuthURL string `json:"auth_url"`
	// The state string for the client to track
	// example: 4fb9-a2c3-987d
	State string `json:"state"`
	// example: google
	Provider string `json:"provider"`
}