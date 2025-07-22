package user

import (
	"forum/models"
	"forum/repository"
	"forum/utils"
)

func (r *UserRepository) Authenticate(login models.UserLogin) (*models.User, error) {
	user, err := r.GetByEmail(login.Email)
	if err != nil {
		return nil, repository.ErrInvalidCredentials
	}

	auth, err := r.GetAuthByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	if !utils.CheckPasswordHash(login.Password, auth.PasswordHash) {
		return nil, repository.ErrInvalidCredentials
	}

	return user, nil
}
