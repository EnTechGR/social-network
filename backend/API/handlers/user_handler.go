package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/repository/user_repository"
	"social-network/utils"
)

// UserHandler handles user profile operations
type UserHandler struct {
	UserRepo  *user_repository.UserRepository
	ImageRepo *repository.ImageRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo *user_repository.UserRepository, imageRepo *repository.ImageRepository) *UserHandler {
	return &UserHandler{
		UserRepo:  userRepo,
		ImageRepo: imageRepo,
	}
}

// GetUserProfile returns another user's profile based on privacy settings
// @Summary      Get user profile by ID
// @Description  Returns profile details of any user. For private profiles, only followers can see full details.
// @Tags         User Profile
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{} "User profile data with posts, followers, and following"
// @Failure      400  {object}  models.ErrorResponse "Missing user ID"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      403  {object}  models.ErrorResponse "Private profile - not a follower"
// @Failure      404  {object}  models.ErrorResponse "User not found"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := middleware.GetCurrentUser(r)
	if currentUser == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL path
	targetUserID := utils.GetLastPathParam(r)
	if targetUserID == "" {
		utils.ErrorResponse(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get target user's profile + avatar data
	targetUser, err := h.UserRepo.GetUserWithAvatar(targetUserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			utils.ErrorResponse(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to get user %s: %v", targetUserID, err)
			utils.ErrorResponse(w, "Failed to retrieve user", http.StatusInternalServerError)
		}
		return
	}

	// Check if viewing own profile
	isOwnProfile := currentUser.ID == targetUserID

	// Check privacy settings
	if targetUser.IsPrivate && !isOwnProfile {
		// Check if current user follows target user
		isFollowing, err := h.UserRepo.IsFollowing(currentUser.ID, targetUserID)
		if err != nil {
			log.Printf("Failed to check follow status: %v", err)
			utils.ErrorResponse(w, "Failed to verify access", http.StatusInternalServerError)
			return
		}

		if !isFollowing {
			// Private profile and not a follower - return limited info
			utils.JSONResponse(w, map[string]interface{}{
				"user_id":    targetUser.ID,
				"nickname":   targetUser.Nickname,
				"first_name": targetUser.FirstName,
				"last_name":  targetUser.LastName,
				"is_private": targetUser.IsPrivate,
				"avatar":     targetUser.Avatar,
				"message":    "This profile is private. Follow to see their content.",
			}, http.StatusForbidden)
			return
		}
	}

	// Get user's posts
	posts, err := h.UserRepo.GetUserPosts(targetUserID)
	if err != nil {
		log.Printf("Failed to get user posts: %v", err)
		posts = []interface{}{} // Return empty array on error
	}

	// Get followers count and list
	followers, err := h.UserRepo.GetFollowers(targetUserID)
	if err != nil {
		log.Printf("Failed to get followers: %v", err)
		followers = []interface{}{}
	}

	// Get following count and list
	following, err := h.UserRepo.GetFollowing(targetUserID)
	if err != nil {
		log.Printf("Failed to get following: %v", err)
		following = []interface{}{}
	}

	// Build response
	response := map[string]interface{}{
		"user":      targetUser,
		"posts":     posts,
		"followers": followers,
		"following": following,
		"counts": map[string]int{
			"posts":     len(posts),
			"followers": len(followers),
			"following": len(following),
		},
		"is_own_profile": isOwnProfile,
	}

	utils.JSONResponse(w, response, http.StatusOK)
}

// GetProfile returns the complete user profile with avatar
// @Summary      Get current user profile
// @Description  Returns the full profile details of the currently authenticated user including avatar information.
// @Tags         User Profile
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  models.UserWithAvatar "Complete user profile data including avatar"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Use the existing repository method that gets user with avatar
	userWithAvatar, err := h.UserRepo.GetUserWithAvatar(user.ID)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, userWithAvatar, http.StatusOK)
}

// TogglePrivacy toggles the user's profile privacy setting
// @Summary      Toggle profile privacy
// @Description  Changes the user's profile from public to private or vice versa. When changing to public, all pending follow requests are automatically cleaned up.
// @Tags         User Profile
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        data body object true "Privacy setting {\"is_private\": true/false}"
// @Success      200  {object}  map[string]interface{} "Returns updated privacy status and message"
// @Failure      400  {object}  models.ErrorResponse "Invalid request body"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/v1/user/privacy [put]
func (h *UserHandler) TogglePrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		IsPrivate bool `json:"is_private"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update privacy setting
	err := h.UserRepo.UpdatePrivacy(user.ID, req.IsPrivate)
	if err != nil {
		log.Printf("Failed to update privacy for user %s: %v", user.ID, err)
		utils.ErrorResponse(w, "Failed to update privacy setting", http.StatusInternalServerError)
		return
	}

	// Update the user in context for consistency
	user.IsPrivate = req.IsPrivate

	message := "Profile is now public"
	if req.IsPrivate {
		message = "Profile is now private"
	}

	log.Printf("[PRIVACY] User %s changed privacy to: isPrivate=%v", user.ID, req.IsPrivate)

	utils.JSONResponse(w, map[string]interface{}{
		"is_private": req.IsPrivate,
		"message":    message,
	}, http.StatusOK)
}

// GetPublicUsers returns all public users for global search.
// @Summary      Get public users
// @Description  Returns all users with public profiles, excluding the authenticated user.
// @Tags         User Profile
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/v1/users/public [get]
func (h *UserHandler) GetPublicUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := middleware.GetCurrentUser(r)
	if currentUser == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := h.UserRepo.GetPublicUsers(currentUser.ID)
	if err != nil {
		log.Printf("Failed to get public users: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	type PublicUser struct {
		ID                  string             `json:"id"`
		Nickname            string             `json:"nickname"`
		Email               string             `json:"email"`
		FirstName           string             `json:"first_name"`
		LastName            string             `json:"last_name"`
		DateOfBirth         string             `json:"date_of_birth"`
		Gender              string             `json:"gender"`
		IsPrivate           bool               `json:"is_private"`
		Avatar              *models.AvatarInfo `json:"avatar,omitempty"`
		AvatarPath          string             `json:"avatar_path,omitempty"`
		AvatarThumbnailPath string             `json:"avatar_thumbnail_path,omitempty"`
		AvatarURL           string             `json:"avatar_url,omitempty"`
		AvatarThumbURL      string             `json:"avatar_thumb_url,omitempty"`
		IsOnline            bool               `json:"is_online"`
	}

	respUsers := make([]PublicUser, 0, len(users))
	for _, u := range users {
		avatarPath := ""
		avatarThumbPath := ""
		if u.Avatar != nil {
			avatarPath = u.Avatar.FilePath
			avatarThumbPath = u.Avatar.ThumbnailPath
		}

		respUsers = append(respUsers, PublicUser{
			ID:                  u.ID,
			Nickname:            u.Nickname,
			Email:               u.Email,
			FirstName:           u.FirstName,
			LastName:            u.LastName,
			DateOfBirth:         u.DateOfBirth.Format(time.RFC3339),
			Gender:              u.Gender,
			IsPrivate:           u.IsPrivate,
			Avatar:              u.Avatar,
			AvatarPath:          avatarPath,
			AvatarThumbnailPath: avatarThumbPath,
			AvatarURL:           avatarPath,
			AvatarThumbURL:      avatarThumbPath,
			IsOnline:            false,
		})
	}

	utils.JSONResponse(w, map[string]interface{}{
		"users": respUsers,
	}, http.StatusOK)
}

// UpdateProfile updates the user's profile information
// @Summary      Update user profile
// @Description  Updates editable user profile fields (first_name, last_name, about_me, gender)
// @Tags         User Profile
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        data body object true "Profile update data"
// @Success      200  {object}  models.User "Updated user profile"
// @Failure      400  {object}  models.ErrorResponse "Invalid request body or validation error"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		AboutMe   string `json:"about_me"`
		Gender    string `json:"gender"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize inputs
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.AboutMe = strings.TrimSpace(req.AboutMe)
	req.Gender = strings.TrimSpace(strings.ToLower(req.Gender))

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" {
		utils.ErrorResponse(w, "First name and last name are required", http.StatusBadRequest)
		return
	}

	// Validate field lengths
	if len(req.FirstName) > 100 {
		utils.ErrorResponse(w, "First name must be at most 100 characters", http.StatusBadRequest)
		return
	}

	if len(req.LastName) > 100 {
		utils.ErrorResponse(w, "Last name must be at most 100 characters", http.StatusBadRequest)
		return
	}

	if len(req.AboutMe) > 500 {
		utils.ErrorResponse(w, "About me must be at most 500 characters", http.StatusBadRequest)
		return
	}

	// Validate gender
	validGenders := map[string]bool{
		"male":              true,
		"female":            true,
		"other":             true,
		"prefer_not_to_say": true,
	}
	if req.Gender != "" && !validGenders[req.Gender] {
		utils.ErrorResponse(w, "Gender must be one of: male, female, other, prefer_not_to_say", http.StatusBadRequest)
		return
	}

	// Update profile
	err := h.UserRepo.UpdateProfile(user.ID, req.FirstName, req.LastName, req.AboutMe, req.Gender)
	if err != nil {
		log.Printf("Failed to update profile for user %s: %v", user.ID, err)
		utils.ErrorResponse(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Get updated user data
	updatedUser, err := h.UserRepo.GetByID(user.ID)
	if err != nil {
		log.Printf("Failed to retrieve updated user: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve updated profile", http.StatusInternalServerError)
		return
	}

	log.Printf("[PROFILE] User %s updated their profile", user.ID)

	utils.JSONResponse(w, updatedUser, http.StatusOK)
}
