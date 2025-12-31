package user_repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"social-network/models"
	"social-network/repository"
	"social-network/utils"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

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

// Create creates a new user with all required fields
// Note: Avatar is now handled separately via ImageRepository.UploadUserAvatar()
func (r *UserRepository) Create(reg models.UserRegistration) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isNicknameTaken(reg.Nickname); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrNicknameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	// ✅ UPDATED: Removed avatar_path and avatar_thumbnail_path
	// Avatars are now stored in images_core and linked via user_avatars
	_, err = tx.Exec(
		`INSERT INTO user (
			user_id, nickname, email, first_name, last_name, 
			date_of_birth, about_me, gender, is_private, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, reg.Nickname, reg.Email, reg.FirstName, reg.LastName,
		reg.DateOfBirth, reg.AboutMe, reg.Gender, reg.IsPrivate, createdAt,
	)
	if err != nil {
		return nil, err
	}

	passwordHash, err := utils.HashPassword(reg.Password)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		"INSERT INTO user_auth (user_id, password_hash) VALUES (?, ?)",
		userID, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.User{
		ID:          userID,
		Nickname:    reg.Nickname,
		Email:       reg.Email,
		FirstName:   reg.FirstName,
		LastName:    reg.LastName,
		DateOfBirth: reg.DateOfBirth,
		AboutMe:     reg.AboutMe,
		Gender:      reg.Gender,
		IsPrivate:   reg.IsPrivate,
		CreatedAt:   createdAt,
	}, nil
}

// CreateOAuthUser creates a new user via OAuth with all required fields
// Note: OAuth avatar URLs are handled via ImageRepository
func (r *UserRepository) CreateOAuthUser(reg models.UserRegistration, provider, providerUserID, avatarURL, accessToken, refreshToken string, tokenExpiresAt time.Time) (*models.User, error) {
	if exists, err := r.isEmailTaken(reg.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrEmailTaken
	}

	if exists, err := r.isNicknameTaken(reg.Nickname); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrNicknameTaken
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := utils.GenerateUUID()
	createdAt := time.Now()

	// ✅ UPDATED: Removed avatar_url field - avatars now managed via images_core
	_, err = tx.Exec(
		`INSERT INTO user (
			user_id, nickname, email, first_name, last_name, 
			date_of_birth, about_me, gender, is_private, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, reg.Nickname, reg.Email, reg.FirstName, reg.LastName,
		reg.DateOfBirth, reg.AboutMe, reg.Gender, reg.IsPrivate, createdAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO oauth_accounts (
			oauth_id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar_url,
			access_token, refresh_token, token_expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		utils.GenerateUUID(),
		userID, provider, providerUserID,
		reg.Nickname, reg.Email,
		avatarURL,
		accessToken, refreshToken, tokenExpiresAt.Format(time.RFC3339),
		createdAt, createdAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Note: OAuth avatar should be downloaded and uploaded via ImageRepository.UploadUserAvatar()
	// after user creation if avatarURL is provided

	return &models.User{
		ID:          userID,
		Nickname:    reg.Nickname,
		Email:       reg.Email,
		FirstName:   reg.FirstName,
		LastName:    reg.LastName,
		DateOfBirth: reg.DateOfBirth,
		AboutMe:     reg.AboutMe,
		Gender:      reg.Gender,
		IsPrivate:   reg.IsPrivate,
		CreatedAt:   createdAt,
	}, nil
}

func (r *UserRepository) isEmailTaken(email string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) isNicknameTaken(nickname string) (bool, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM user WHERE nickname = ?", nickname).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) IsProviderLinked(userID, provider string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM oauth_accounts
		WHERE user_id = ? AND provider = ?
	`, userID, provider).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) LinkOAuthProvider(userID, provider, providerUserID, accessToken, refreshToken string, expiresAt time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO oauth_accounts (
			oauth_id, user_id, provider, provider_user_id,
			access_token, refresh_token, token_expires_at,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(provider, provider_user_id)
		DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_expires_at = excluded.token_expires_at,
			updated_at = excluded.updated_at
	`,
		utils.GenerateUUID(), userID, provider, providerUserID,
		accessToken, refreshToken, expiresAt.Format(time.RFC3339),
		time.Now(), time.Now(),
	)
	return err
}

// DeleteByID deletes a user by ID (used for cleanup on registration failure)
func (r *UserRepository) DeleteByID(userID string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete user_auth first due to foreign key
	_, err = tx.Exec("DELETE FROM user_auth WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Delete user
	_, err = tx.Exec("DELETE FROM user WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

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

// UpdatePrivacy updates the user's privacy setting
// CRITICAL SECURITY: When changing to public, this MUST clean up pending follow requests
// to prevent the security vulnerability of auto-accepting unapproved follow requests
func (r *UserRepository) UpdatePrivacy(userID string, isPrivate bool) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update user privacy setting
	result, err := tx.Exec(`
		UPDATE user 
		SET is_private = ? 
		WHERE user_id = ?
	`, isPrivate, userID)

	if err != nil {
		return fmt.Errorf("failed to update privacy: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// 2. CRITICAL SECURITY CHECK: Clean up pending follow requests when going public
	// This prevents the security vulnerability where:
	// - User has private profile with pending (unapproved) follow requests
	// - User changes to public profile
	// - Pending requests should NOT auto-accept (user never approved them)
	// - Solution: Delete all pending requests when profile becomes public
	if !isPrivate {
		result, err := tx.Exec(`
			DELETE FROM follow_relationships 
			WHERE followee_id = ? AND status = 'pending'
		`, userID)

		if err != nil {
			return fmt.Errorf("failed to clean up pending requests: %w", err)
		}

		deleted, _ := result.RowsAffected()
		if deleted > 0 {
			log.Printf("[SECURITY] User %s → PUBLIC: Cleaned up %d pending follow requests", userID, deleted)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[PRIVACY] User %s privacy changed to: isPrivate=%v", userID, isPrivate)
	return nil
}

// UpdateProfile updates editable user profile fields
func (r *UserRepository) UpdateProfile(userID, firstName, lastName, aboutMe, gender string) error {
	result, err := r.DB.Exec(`
		UPDATE user 
		SET first_name = ?,
		    last_name = ?,
		    about_me = ?,
		    gender = ?
		WHERE user_id = ?
	`, firstName, lastName, aboutMe, gender, userID)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetCompleteProfile returns the user with all their information including avatar
// This is useful for profile pages where we need all user data
func (r *UserRepository) GetCompleteProfile(userID string) (*models.UserWithAvatar, error) {
	var user models.User
	var createdAt sql.NullTime

	// Get basic user information
	err := r.DB.QueryRow(`
		SELECT user_id, nickname, email, first_name, last_name, date_of_birth, 
		       about_me, gender, is_private, created_at 
		FROM user 
		WHERE user_id = ?
	`, userID).Scan(
		&user.ID, &user.Nickname, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.AboutMe, &user.Gender, &user.IsPrivate, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt = createdAt.Time

	// Create response with user data
	// Note: Avatar will be fetched separately by the handler using ImageRepository
	profile := &models.UserWithAvatar{
		User: user,
	}

	return profile, nil
}
