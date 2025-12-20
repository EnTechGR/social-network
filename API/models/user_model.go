package models

import "time"

// User represents a forum user
// swagger:model User
type User struct {
	// The unique identifier for the user.
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	ID string `json:"id"`
	// The unique username of the user.
	// example: johndoe
	Username string `json:"username"`
	// The email address of the user.
	// example: john.doe@example.com
	Email string `json:"email"`
	// The first name of the user.
	// example: John
	FirstName string `json:"first_name"` 
	// The last name of the user.
	// example: Doe
	LastName string `json:"last_name"` 
	// The age of the user.
	// example: 35
	Age int `json:"age"` 
	// The gender of the user (e.g., Male, Female, Other).
	// example: Male
	Gender string `json:"gender"` 
	// The timestamp when the user account was created.
	// example: 2025-01-01T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// UserRegistration is used for registration requests
// swagger:model UserRegistration
type UserRegistration struct {
	// The desired unique username.
	// required: true
	// min length: 4
	// example: new_forum_user
	Username string `json:"username"`
	// The user's email address. Must be unique and valid.
	// required: true
	// example: register@example.com
	Email string `json:"email"`
	// The user's chosen password.
	// required: true
	// min length: 8
	// example: secureP@ss123
	Password string `json:"password"`
	// The user's first name.
	// required: true
	// example: Jane
	FirstName string `json:"first_name"`
	// The user's last name.
	// required: true
	// example: Smith
	LastName string `json:"last_name"`
	// The user's age. Must be between 13 and 120.
	// required: true
	// minimum: 13
	// maximum: 120
	// example: 24
	Age int `json:"age"`
	// The user's gender.
	// required: true
	// example: Female
	Gender string `json:"gender"`
}

// UserLogin is used for login requests
// swagger:model UserLogin
type UserLogin struct {
	// The user's username OR email address.
	// required: true
	// example: johndoe OR john.doe@example.com
	Login string `json:"login"`
	// The user's password.
	// required: true
	// example: secureP@ss123
	Password string `json:"password"`
}

// UserProfile represents extended user information including OAuth accounts
// swagger:model UserProfile
type UserProfile struct {
	// The core details of the user.
	User User `json:"user"`
	// A list of external OAuth accounts linked to this user (e.g., Google, GitHub).
	// items:
	//   "$ref": "#/definitions/OAuthAccount" 
	OAuthAccounts []OAuthAccount `json:"oauth_accounts"`
	// Indicates whether the user has a local password set (true) or is OAuth-only (false).
	HasPassword bool `json:"has_password"`
}

// LoginResponse is the response after successful login
// swagger:model LoginResponse
type LoginResponse struct {
	// The details of the successfully logged-in user.
	User User `json:"user"`
	// The session ID (often set as a cookie, but included here for context).
	// example: d2e9f8a7-b6c5-d4e3-f2a1-b0c9d8e7f6a5
	SessionID string `json:"session_id"`
	// The CSRF token required for subsequent protected POST/PUT/DELETE requests.
	// example: XyZ123AbC456DeF789GhI0JkL
	CSRFToken string `json:"csrf_token"`
}

// UserAuth contains user authentication information (Internal Use Only)
// swagger:ignore
type UserAuth struct {
	UserID       string `json:"-"`
	PasswordHash string `json:"-"`
}
