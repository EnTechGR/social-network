package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/repository/session"
	"forum/repository/user"
	"forum/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	UserRepo    *user.UserRepository
	SessionRepo *session.SessionRepository
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *user.UserRepository, sessionRepo *session.SessionRepository) *AuthHandler {
    return &AuthHandler{
        UserRepo:    userRepo,
        SessionRepo: sessionRepo,
    }
}

// Register handles user registration with all required fields
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg models.UserRegistration
	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim and normalize inputs
	reg.Nickname = strings.TrimSpace(reg.Nickname)
	reg.Email = strings.TrimSpace(strings.ToLower(reg.Email))
	reg.Password = strings.TrimSpace(reg.Password)
	reg.FirstName = strings.TrimSpace(reg.FirstName)
	reg.LastName = strings.TrimSpace(reg.LastName)
	reg.Gender = strings.TrimSpace(strings.ToLower(reg.Gender))
	// ✅ ADDED: Normalize new profile fields
	reg.AboutMe = strings.TrimSpace(reg.AboutMe)
	reg.AvatarURL = strings.TrimSpace(reg.AvatarURL)

	// ✅ UPDATED: Validation for required fields
	if reg.Nickname == "" || reg.Email == "" || reg.Password == "" ||
		reg.FirstName == "" || reg.LastName == "" || reg.DateOfBirth.IsZero() || reg.Gender == "" {
		utils.ErrorResponse(w, "All fields are required: nickname, email, password, first_name, last_name, date_of_birth, gender", http.StatusBadRequest)
		return
	}

	// Nickname: 1-50 chars
	if len(reg.Nickname) < 1 || len(reg.Nickname) > 50 {
		utils.ErrorResponse(w, "Nickname must be between 1 and 50 characters", http.StatusBadRequest)
		return
	}

	// Email validation
	cleanEmail, err := utils.ValidateEmail(reg.Email)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	reg.Email = cleanEmail

	// Password strength
	if !utils.IsStrongPassword(reg.Password) {
		utils.ErrorResponse(w, "Password must be at least 8 characters, with at least one letter and one digit", http.StatusBadRequest)
		return
	}

	// ✅ UPDATED: Validate Age via DateOfBirth (13-120 years)
	now := time.Now()
	age := now.Year() - reg.DateOfBirth.Year()
	if now.YearDay() < reg.DateOfBirth.YearDay() {
		age--
	}

	if age < 13 || age > 120 {
		utils.ErrorResponse(w, "You must be between 13 and 120 years old", http.StatusBadRequest)
		return
	}

	// Validate Gender
	validGenders := map[string]bool{
		"male":              true,
		"female":            true,
		"other":             true,
		"prefer_not_to_say": true,
	}
	if !validGenders[reg.Gender] {
		utils.ErrorResponse(w, "Gender must be one of: male, female, other, prefer_not_to_say", http.StatusBadRequest)
		return
	}

	// ✅ ADDED: Validate AboutMe length
	if len(reg.AboutMe) > 500 {
		utils.ErrorResponse(w, "About me must be under 500 characters", http.StatusBadRequest)
		return
	}

	// ✅ ADDED: Set default avatar if empty
	if reg.AvatarURL == "" {
		reg.AvatarURL = "/static/avatars/defaults/default-avatar.png"
	}

	// Create user in DB (Repo now handles avatar_url, about_me, and is_private)
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
// @Router       /forum/api/session/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var login models.UserLogin
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// ✅ UPDATED: Trim the login field (now accepts username OR email)
	login.Login = strings.TrimSpace(login.Login)
	login.Password = strings.TrimSpace(login.Password)

	// ✅ UPDATED: Validate request (now checks login instead of email)
	if login.Login == "" || login.Password == "" {
		utils.ErrorResponse(w, "Username/email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user (now supports both username and email)
	user, err := h.UserRepo.Authenticate(login)
	if err != nil {
		if err == repository.ErrInvalidCredentials {
			// ✅ UPDATED: Generic error message for security
			utils.ErrorResponse(w, "Invalid username/email or password", http.StatusUnauthorized)
		} else {
			log.Printf("Authentication error: %v", err)
			utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create session after successful authentication
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
// @Description  Deletes the current user session from the database and clears the session and CSRF cookies.
// @Tags         Authentication
// @Produce      plain
// @Success      200 "Successfully logged out (session cookie cleared)."
// @Router       /forum/api/session/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// If no cookie, nothing to do
		w.WriteHeader(http.StatusOK)
		return
	}

	// Delete the session from database
	err = h.SessionRepo.Delete(cookie.Value)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
		// Continue with clearing cookie even if DB delete fails
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // true in production
		SameSite: http.SameSiteLaxMode,
	})

	// Clear the CSRF token cookie (if used on the client)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false, // false if your frontend JS needs to read it
		Secure:   false, // true in production
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

// VerifySession handles session verification
// @Summary      Verify current session status
// @Description  Checks if the session cookie is valid, active, and not expired. Returns user data if session is valid.
// @Tags         Authentication
// @Produce      json
// @Success      200  {object}  object "Session is valid."
// @Success      200  {object}  object{user=models.User,csrf_token=string}
// @Failure      401  {object}  models.ErrorResponse "Unauthorized: Session cookie not found, invalid, or expired."
// @Failure      500  {object}  models.ErrorResponse "Internal server error."
// @Router       /forum/api/session/verify [get]
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

	// Check if session is expired
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

	// Return user data + csrf token
	utils.JSONResponse(w, struct {
		User      *models.User `json:"user"`
		CSRFToken string       `json:"csrf_token"`
	}{
		User:      user,
		CSRFToken: session.CSRFToken,
	}, http.StatusOK)
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
		Secure:   false, // true in prod
		SameSite: http.SameSiteLaxMode,
	})

	return session, nil
}

// LogoutAll handles logout from all devices
// @Summary      Log out from all devices
// @Description  Deletes all active sessions for the currently authenticated user and clears the current session cookie.
// @Tags         Authentication
// @Produce      plain
// @Success      200 "Successfully logged out from all devices."
// @Failure      401  "Unauthorized: User not authenticated." // NOTE: We cannot use models.ErrorResponse here unless the handler uses utils.ErrorResponse()
// @Failure      500  {object}  models.ErrorResponse "Internal server error."
// @Router       /forum/api/session/logout-all [post]
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

	// Delete all sessions for this user
	err := h.SessionRepo.DeleteAllUserSessions(user.ID)
	if err != nil {
		log.Printf("Failed to delete all user sessions: %v", err)
		http.Error(w, "Failed to logout from all devices", http.StatusInternalServerError)
		return
	}

	// Clear current session cookie
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
// @Description  Returns the profile data for the user associated with the active session.
// @Tags         User Profile
// @Produce      json
// @Success      200  {object}  models.User "Successful response."
// @Failure      401  "Unauthorized: User not authenticated."
// @Router       /forum/api/user/profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, user, http.StatusOK)
}



