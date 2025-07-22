package models

// UserRegistration is used for registration requests
type UserRegistration struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}


// UserLogin is used for login requests
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// OAuthLoginRequest represents OAuth login initiation
type OAuthLoginRequest struct {
	Provider    string `json:"provider" binding:"required"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// OAuthCallbackRequest represents OAuth callback data
type OAuthCallbackRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// AccountLinkRequest represents a request to link an OAuth account
type AccountLinkRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// AccountUnlinkRequest represents a request to unlink an OAuth account
type AccountUnlinkRequest struct {
	Provider string `json:"provider" binding:"required"`
}