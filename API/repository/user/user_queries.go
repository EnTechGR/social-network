package user

import (
	"database/sql"
	"strings"
	
	"social-network/models"
	"social-network/repository"
)

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	// ✅ UPDATED: Removed avatar_path and avatar_thumbnail_path from SELECT
	// Avatars are now retrieved separately using ImageRepository.GetUserAvatar()
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
			about_me, gender, is_private, created_at 
		FROM user WHERE email = ?`,
		email,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	// ✅ UPDATED: Removed avatar fields
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
			about_me, gender, is_private, created_at 
		FROM user WHERE user_id = ?`,
		id,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByNickname retrieves a user by nickname
func (r *UserRepository) GetByNickname(nickname string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	// ✅ UPDATED: Removed avatar fields
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
			about_me, gender, is_private, created_at 
		FROM user WHERE nickname = ?`,
		nickname,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetByEmailOrNickname retrieves a user by either email or nickname
func (r *UserRepository) GetByEmailOrNickname(login string) (*models.User, error) {
	var user models.User
	var createdAt sql.NullTime

	// ✅ UPDATED: Removed avatar fields
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
			about_me, gender, is_private, created_at 
		FROM user WHERE email = ? OR nickname = ?`,
		login, login,
	).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time
	return &user, nil
}

// GetAuthByUserID retrieves authentication data for a user
func (r *UserRepository) GetAuthByUserID(userID string) (*models.UserAuth, error) {
	var auth models.UserAuth

	err := r.DB.QueryRow(
		"SELECT user_id, password_hash FROM user_auth WHERE user_id = ?",
		userID,
	).Scan(&auth.UserID, &auth.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &auth, nil
}

// GetUserWithAvatar retrieves a user along with their avatar information
// ✅ NEW METHOD: Uses the new centralized image management system
func (r *UserRepository) GetUserWithAvatar(userID string) (*models.UserWithAvatar, error) {
	var userWithAvatar models.UserWithAvatar
	var createdAt sql.NullTime

	// Get user data
	err := r.DB.QueryRow(
		`SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
			about_me, gender, is_private, created_at 
		FROM user WHERE user_id = ?`,
		userID,
	).Scan(
		&userWithAvatar.ID, &userWithAvatar.Nickname, &userWithAvatar.Email,
		&userWithAvatar.FirstName, &userWithAvatar.LastName, &userWithAvatar.DateOfBirth,
		&userWithAvatar.AboutMe, &userWithAvatar.Gender, &userWithAvatar.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	userWithAvatar.CreatedAt = createdAt.Time

	// Get avatar data from v_user_avatars view
	var avatar models.AvatarInfo
	err = r.DB.QueryRow(
		`SELECT image_id, file_path, thumbnail_path, mime_type, set_at 
		FROM v_user_avatars WHERE user_id = ?`,
		userID,
	).Scan(&avatar.ImageID, &avatar.FilePath, &avatar.ThumbnailPath, &avatar.MimeType, &avatar.SetAt)

	if err == nil {
		// Avatar found
		userWithAvatar.Avatar = &avatar
	} else if err != sql.ErrNoRows {
		// Real error (not just missing avatar)
		return nil, err
	}
	// If err == sql.ErrNoRows, avatar is nil (user has no avatar), which is fine

	return &userWithAvatar, nil
}

// GetMultipleUsersWithAvatars retrieves multiple users with their avatars
// ✅ NEW METHOD: Efficient batch retrieval for user lists
func (r *UserRepository) GetMultipleUsersWithAvatars(userIDs []string) ([]models.UserWithAvatar, error) {
	if len(userIDs) == 0 {
		return []models.UserWithAvatar{}, nil
	}

	// Build query with placeholders
	query := `
		SELECT 
			u.user_id, u.nickname, u.email, u.first_name, u.last_name, u.date_of_birth,
			u.about_me, u.gender, u.is_private, u.created_at,
			va.image_id, va.file_path, va.thumbnail_path, va.mime_type, va.set_at
		FROM user u
		LEFT JOIN v_user_avatars va ON u.user_id = va.user_id
		WHERE u.user_id IN (?` + strings.Repeat(",?", len(userIDs)-1) + `)`

	// Convert userIDs to []interface{} for query
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserWithAvatar
	for rows.Next() {
		var user models.UserWithAvatar
		var createdAt sql.NullTime
		var avatar models.AvatarInfo
		var imageID, filePath, thumbnailPath, mimeType, setAt sql.NullString

		err := rows.Scan(
			&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
			&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
			&imageID, &filePath, &thumbnailPath, &mimeType, &setAt,
		)
		if err != nil {
			return nil, err
		}

		user.CreatedAt = createdAt.Time

		// Set avatar if exists
		if imageID.Valid {
			avatar.ImageID = imageID.String
			avatar.FilePath = filePath.String
			avatar.ThumbnailPath = thumbnailPath.String
			avatar.MimeType = mimeType.String
			avatar.SetAt = setAt.String
			user.Avatar = &avatar
		}

		users = append(users, user)
	}

	return users, nil
}