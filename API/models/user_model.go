package models

import "time"

// User represents a forum user
// swagger:model User
type User struct {
    // The unique identifier for the user.
    // example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
    ID string `json:"id"`
    // The unique nickname of the user.
    // example: johndoe
    Nickname string `json:"nickname"`
    // The email address of the user.
    // example: john.doe@example.com
    Email string `json:"email"`
    // The first name of the user.
    // example: John
    FirstName string `json:"first_name"`
    // The last name of the user.
    // example: Doe
    LastName string `json:"last_name"`
    // The date of birth of the user.
    // example: 1990-01-01
    DateOfBirth time.Time `json:"date_of_birth"`
    // The URL for the user's avatar.
    AvatarURL string `json:"avatar_url"`
    // A short bio or description of the user.
    AboutMe string `json:"about_me"`
    // The gender of the user (male, female, other, prefer_not_to_say).
    // example: male
    Gender string `json:"gender"`
    // Indicates if the profile is private.
    IsPrivate bool `json:"is_private"`
    // The timestamp when the user account was created.
    // example: 2025-01-01T10:00:00Z
    CreatedAt time.Time `json:"created_at"`
}

// UserRegistration is used for registration requests
// swagger:model UserRegistration
type UserRegistration struct {
    // The desired unique nickname.
    // required: true
    // min length: 1
    // max length: 50
    // example: new_forum_user
    Nickname string `json:"nickname"`
    // The user's email address. Must be unique and valid.
    // required: true
    // max length: 100
    // example: register@example.com
    Email string `json:"email"`
    // The user's chosen password.
    // required: true
    // min length: 8
    // max length: 255
    // example: secureP@ss123
    Password string `json:"password"`
    // The user's first name.
    // required: true
    // max length: 100
    // example: Jane
    FirstName string `json:"first_name"`
    // The user's last name.
    // required: true
    // max length: 100
    // example: Smith
    LastName string `json:"last_name"`
    // The user's date of birth.
    // required: true
    // example: 2000-01-01
    DateOfBirth time.Time `json:"date_of_birth"`
    // The user's gender.
    // required: true
    // example: female
    Gender string `json:"gender"`
}

// UserLogin is used for login requests
// swagger:model UserLogin
type UserLogin struct {
    // The user's nickname OR email address.
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
    OAuthAccounts []OAuthAccount `json:"oauth_accounts"`
    // Indicates whether the user has a local password set (true) or is OAuth-only (false).
    HasPassword bool `json:"has_password"`
}

// LoginResponse is the response after successful login
// swagger:model LoginResponse
type LoginResponse struct {
    // The details of the successfully logged-in user.
    User User `json:"user"`
    // The session ID.
    // example: d2e9f8a7-b6c5-d4e3-f2a1-b0c9d8e7f6a5
    SessionID string `json:"session_id"`
    // The CSRF token required for subsequent protected requests.
    // example: XyZ123AbC456DeF789GhI0JkL
    CSRFToken string `json:"csrf_token"`
}

// UserAuth contains user authentication information (Internal Use Only)
// swagger:ignore
type UserAuth struct {
    UserID       string `json:"-"`
    PasswordHash string `json:"-"`
}