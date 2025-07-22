package models

// UserAuth contains user authentication information
type UserAuth struct {
	UserID       string `json:"-"`
	PasswordHash string `json:"-"`
}
