package models

// OAuthLoginRequest represents the initial request to start an OAuth flow
// swagger:model OAuthLoginRequest
type OAuthLoginRequest struct {
	// The OAuth provider to use
	// example: google
	// required: true
	Provider string `json:"provider" binding:"required"`
	// Optional URL to redirect to after successful login
	// example: https://my-app.com/dashboard
	RedirectURL string `json:"redirect_url,omitempty"`
}

// OAuthCallbackRequest represents the data sent back from the OAuth provider
// swagger:model OAuthCallbackRequest
type OAuthCallbackRequest struct {
	// The OAuth provider that sent the callback
	// example: github
	// required: true
	Provider string `json:"provider" binding:"required"`
	// The authorization code provided by the OAuth service
	// example: 4/0AfgeXvv...
	// required: true
	Code string `json:"code" binding:"required"`
	// The state string for CSRF protection
	// example: xyz123
	// required: true
	State string `json:"state" binding:"required"`
}

// AccountLinkRequest represents a request to link an OAuth provider to an existing account
// swagger:model AccountLinkRequest
type AccountLinkRequest struct {
	// The OAuth provider to link
	// example: discord
	// required: true
	Provider string `json:"provider" binding:"required"`
	// The authorization code from the provider
	// required: true
	Code string `json:"code" binding:"required"`
	// The state string for security verification
	// required: true
	State string `json:"state" binding:"required"`
}

// AccountUnlinkRequest represents a request to disconnect an OAuth provider
// swagger:model AccountUnlinkRequest
type AccountUnlinkRequest struct {
	// The provider to disconnect (e.g., "google")
	// example: google
	// required: true
	Provider string `json:"provider" binding:"required"`
}