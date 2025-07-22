package repository

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email is already taken")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrOAuthAccountNotFound = errors.New("oauth account not found")
	ErrOAuthStateNotFound   = errors.New("oauth state not found")
	ErrOAuthStateExpired    = errors.New("oauth state expired")
	ErrOAuthAccountExists   = errors.New("oauth account already exists")
)
