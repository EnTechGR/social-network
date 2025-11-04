package models

import "time"

// User represents a forum user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`  // ✅ Already present
	LastName  string    `json:"last_name"`   // ✅ Already present
	Age       int       `json:"age"`         // ✅ Already present
	Gender    string    `json:"gender"`      // ✅ Already present
	CreatedAt time.Time `json:"created_at"`
}

// UserRegistration is used for registration requests
type UserRegistration struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"` // ✅ ADD THIS
	LastName  string `json:"last_name" binding:"required"`  // ✅ ADD THIS
	Age       int    `json:"age" binding:"required,min=13,max=120"` // ✅ ADD THIS
	Gender    string `json:"gender" binding:"required"` // ✅ ADD THIS
}


// UserLogin is used for login requests
type UserLogin struct {
	Login    string `json:"login" binding:"required"`    // ✅ CHANGE: Was "Email", now "Login" (accepts username OR email)
	Password string `json:"password" binding:"required"`
}

// UserProfile represents extended user information including OAuth accounts
type UserProfile struct {
	User          User           `json:"user"`
	OAuthAccounts []OAuthAccount `json:"oauth_accounts"`
	HasPassword   bool           `json:"has_password"` // Whether user has a password (for mixed auth)
}

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