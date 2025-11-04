package user

import (
	"strings"

	"forum/models"
	"forum/repository"
	"forum/utils"
)

// Authenticate verifies user credentials using username OR email
func (r *UserRepository) Authenticate(login models.UserLogin) (*models.User, error) {
	// ✅ CHANGED: Now uses GetByEmailOrUsername instead of GetByEmail
	// This allows login with either username or email
	user, err := r.GetByEmailOrUsername(strings.TrimSpace(login.Login))
	if err != nil {
		return nil, repository.ErrInvalidCredentials
	}

	auth, err := r.GetAuthByUserID(user.ID)
	if err != nil {
		return nil, repository.ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(login.Password, auth.PasswordHash) {
		return nil, repository.ErrInvalidCredentials
	}

	return user, nil
}