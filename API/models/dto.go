package models

// LoginResponse is the response after successful login
type LoginResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
	CSRFToken string `json:"csrf_token"`
}

// OAuthLoginResponse is the response for OAuth login initiation
type OAuthLoginResponse struct {
	AuthURL   string `json:"auth_url"`
	State     string `json:"state"`
	Provider  string `json:"provider"`
}

// UserProfile represents extended user information including OAuth accounts
type UserProfile struct {
	User          User           `json:"user"`
	OAuthAccounts []OAuthAccount `json:"oauth_accounts"`
	HasPassword   bool           `json:"has_password"` // Whether user has a password (for mixed auth)
}