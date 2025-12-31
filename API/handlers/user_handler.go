package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"social-network/middleware"
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
// @Router       /api/user/privacy [put]
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
// @Router       /api/user/profile [put]
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
