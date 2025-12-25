package user

import (
	"strings"

	"forum/models"
	"forum/repository"
	"forum/utils"
)

// Authenticate verifies user credentials using nickname OR email
func (r *UserRepository) Authenticate(login models.UserLogin) (*models.User, error) {
	// ✅ UPDATED: Uses GetByEmailOrNickname to match the new schema
	// This allows login with either the unique nickname or the email address
	user, err := r.GetByEmailOrNickname(strings.TrimSpace(login.Login))
	if err != nil {
		return nil, repository.ErrInvalidCredentials
	}

	auth, err := r.GetAuthByUserID(user.ID)
	if err != nil {
		// Even if the user exists but auth record is missing, 
		// we return ErrInvalidCredentials for security (don't leak existence)
		return nil, repository.ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(login.Password, auth.PasswordHash) {
		return nil, repository.ErrInvalidCredentials
	}

	return user, nil
}