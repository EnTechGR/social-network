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
	"social-network/repository/session"
	"social-network/repository/user"
	"social-network/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	UserRepo    *user.UserRepository
	SessionRepo *session.SessionRepository
	ImageRepo   *repository.ImageRepository // ✅ ADDED: Image repository for avatar handling
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *user.UserRepository, sessionRepo *session.SessionRepository, imageRepo *repository.ImageRepository) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		ImageRepo:   imageRepo,
	}
}

// Register handles user registration with optional avatar upload
// @Summary      Register a new user
// @Description  Creates a new user account. Accepts multipart/form-data for avatar uploads.
// @Tags         Authentication
// @Accept       multipart/form-data
// @Produce      json
// @Param        email          formData  string  true   "User email"
// @Param        password       formData  string  true   "User password (min 8 chars, 1 letter, 1 digit)"
// @Param        first_name     formData  string  true   "User first name"
// @Param        last_name      formData  string  true   "User last name"
// @Param        date_of_birth  formData  string  true   "User date of birth (YYYY-MM-DD)"
// @Param        gender         formData  string  true   "male, female, other, or prefer_not_to_say"
// @Param        nickname       formData  string  false  "Optional: defaults to email prefix"
// @Param        about_me       formData  string  false  "Optional: Max 500 chars"
// @Param        is_private     formData  string  false  "Set to 'true' or '1' for private profile"
// @Param        avatar         formData  file    false  "JPEG/PNG/GIF, max 5MB"
// @Success      201  {object}  models.LoginResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      409  {object}  models.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB to accommodate avatar uploads)
	const maxFormSize = 10 << 20 // 10MB
	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		utils.ErrorResponse(w, "Invalid form data or file too large", http.StatusBadRequest)
		return
	}

	// Extract form fields
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := strings.TrimSpace(r.FormValue("password"))
	firstName := strings.TrimSpace(r.FormValue("first_name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	dobStr := strings.TrimSpace(r.FormValue("date_of_birth"))
	gender := strings.TrimSpace(strings.ToLower(r.FormValue("gender")))
	nickname := strings.TrimSpace(r.FormValue("nickname"))
	aboutMe := strings.TrimSpace(r.FormValue("about_me"))
	isPrivateStr := strings.TrimSpace(r.FormValue("is_private"))

	// Parse date of birth
	dateOfBirth, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		utils.ErrorResponse(w, "Invalid date_of_birth format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Parse is_private (defaults to false)
	isPrivate := false
	if isPrivateStr == "true" || isPrivateStr == "1" {
		isPrivate = true
	}

	// Validate required fields
	if email == "" || password == "" || firstName == "" || lastName == "" || dobStr == "" || gender == "" {
		utils.ErrorResponse(w, "Required fields: email, password, first_name, last_name, date_of_birth, gender", http.StatusBadRequest)
		return
	}

	// Set default nickname if not provided
	if nickname == "" {
		nickname = strings.Split(email, "@")[0]
	}

	// Validate nickname length
	if len(nickname) > 50 {
		utils.ErrorResponse(w, "Nickname must be at most 50 characters", http.StatusBadRequest)
		return
	}

	// Email validation
	cleanEmail, err := utils.ValidateEmail(email)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	email = cleanEmail

	// Password strength
	if !utils.IsStrongPassword(password) {
		utils.ErrorResponse(w, "Password must be at least 8 characters, with at least one letter and one digit", http.StatusBadRequest)
		return
	}

	// Validate age (13-120 years)
	now := time.Now()
	age := now.Year() - dateOfBirth.Year()
	if now.YearDay() < dateOfBirth.YearDay() {
		age--
	}

	if age < 13 || age > 120 {
		utils.ErrorResponse(w, "You must be between 13 and 120 years old", http.StatusBadRequest)
		return
	}

	// Validate gender
	validGenders := map[string]bool{
		"male":              true,
		"female":            true,
		"other":             true,
		"prefer_not_to_say": true,
	}
	if !validGenders[gender] {
		utils.ErrorResponse(w, "Gender must be one of: male, female, other, prefer_not_to_say", http.StatusBadRequest)
		return
	}

	// Validate about_me length
	if len(aboutMe) > 500 {
		utils.ErrorResponse(w, "About me must be under 500 characters", http.StatusBadRequest)
		return
	}

	// Create user registration model
	reg := models.UserRegistration{
		Nickname:    nickname,
		Email:       email,
		Password:    password,
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: dateOfBirth,
		Gender:      gender,
		AboutMe:     aboutMe,
		IsPrivate:   isPrivate,
	}

	// Create user in DB (without avatar - that's handled separately now)
	user, err := h.UserRepo.Create(reg)
	if err != nil {
		switch err {
		case repository.ErrEmailTaken:
			utils.ErrorResponse(w, "Email is already taken", http.StatusConflict)
		case repository.ErrNicknameTaken:
			utils.ErrorResponse(w, "Nickname is already taken", http.StatusConflict)
		default:
			log.Printf("Failed to create user: %v", err)
			utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// ✅ Handle avatar upload if provided
	file, header, err := r.FormFile("avatar")
	if err == nil {
		// Avatar was provided
		defer file.Close()

		// Validate avatar size (max 5MB)
		const maxAvatarSize = 5 << 20 // 5MB
		if header.Size > maxAvatarSize {
			// Delete the created user since avatar upload failed
			h.UserRepo.DeleteByID(user.ID)
			utils.ErrorResponse(w, "Avatar exceeds 5 MB limit", http.StatusBadRequest)
			return
		}

		// Upload avatar using ImageRepository
		if err := h.ImageRepo.UploadUserAvatar(file, header, user.ID); err != nil {
			log.Printf("Failed to upload avatar for user %s: %v", user.ID, err)
			// Don't fail registration, just log the error
			// User can upload avatar later via profile update
			log.Printf("Warning: User %s created but avatar upload failed", user.ID)
		}
	}
	// If no avatar provided (err != nil), that's fine - avatar is optional

	// Create session
	session, err := h.createUserSession(w, r, user)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, models.LoginResponse{
		User:      *user,
		SessionID: session.SessionID,
		CSRFToken: session.CSRFToken,
	}, http.StatusCreated)
}

// Login handles user login with username OR email
// @Summary      Log in a user
// @Description  Authenticates a user using either username or email and password, then creates a session cookie.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        data body models.UserLogin true "User login credentials"
// @Success      200  {object}  models.LoginResponse "User successfully logged in."
// @Failure      400  {object}  models.ErrorResponse "Invalid request body or missing credentials."
// @Failure      401  {object}  models.ErrorResponse "Unauthorized: Invalid username/email or password."
// @Failure      500  {object}  models.ErrorResponse "Internal server error."
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var login models.UserLogin
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	login.Login = strings.TrimSpace(login.Login)
	login.Password = strings.TrimSpace(login.Password)

	if login.Login == "" || login.Password == "" {
		utils.ErrorResponse(w, "Username/email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.Authenticate(login)
	if err != nil {
		if err == repository.ErrInvalidCredentials {
			utils.ErrorResponse(w, "Invalid username/email or password", http.StatusUnauthorized)
		} else {
			log.Printf("Authentication error: %v", err)
			utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	session, err := h.createUserSession(w, r, user)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, models.LoginResponse{
		User:      *user,
		SessionID: session.SessionID,
		CSRFToken: session.CSRFToken,
	}, http.StatusOK)
}

// Logout handles user logout
// @Summary      Log out a user
// @Description  Invalidates the user's session and clears authentication cookies.
// @Tags         Authentication
// @Produce      json
// @Success      200  {string}  string "Successfully logged out"
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.SessionRepo.Delete(cookie.Value)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

// VerifySession handles session verification
// @Summary      Verify current session
// @Description  Checks if the session cookie is valid and returns the current user's info and a fresh CSRF token.
// @Tags         Authentication
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  SessionVerifyResponse "Current user and CSRF token"
// @Failure      401  {object}  models.ErrorResponse
// @Router       /api/auth/verify [get]
func (h *AuthHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	session, err := h.SessionRepo.GetBySessionID(sessionCookie.Value)
	if err != nil {
		http.Error(w, "Session invalid or expired", http.StatusUnauthorized)
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		h.SessionRepo.Delete(session.SessionID)
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	user, err := h.UserRepo.GetByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, struct {
		User      *models.User `json:"user"`
		CSRFToken string       `json:"csrf_token"`
	}{
		User:      user,
		CSRFToken: session.CSRFToken,
	}, http.StatusOK)
}

// SessionVerifyResponse represents the specific response for session verification
type SessionVerifyResponse struct {
    User      models.User `json:"user"`
    CSRFToken string      `json:"csrf_token"`
}

// createUserSession creates a session and sets the session cookie
func (h *AuthHandler) createUserSession(w http.ResponseWriter, r *http.Request, user *models.User) (*models.Session, error) {
	csrfToken := utils.GenerateCSRFToken()
	session, err := h.SessionRepo.Create(user.ID, r.RemoteAddr, csrfToken)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.SessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return session, nil
}

// LogoutAll handles logout from all devices
// @Summary      Log out from all devices
// @Description  Deletes all active sessions for the current user across all devices.
// @Tags         Authentication
// @Security     CookieAuth
// @Produce      json
// @Success      200  {string}  string "Successfully logged out from all devices"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Router       /api/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.SessionRepo.DeleteAllUserSessions(user.ID)
	if err != nil {
		log.Printf("Failed to delete all user sessions: %v", err)
		http.Error(w, "Failed to logout from all devices", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

// GetProfile returns the current user's profile
// @Summary      Get current user profile
// @Description  Returns the full profile details of the currently authenticated user.
// @Tags         Authentication
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  models.User "User profile data"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Router       /api/auth/profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, user, http.StatusOK)
}