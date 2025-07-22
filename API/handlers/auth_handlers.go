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

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reg models.UserRegistration
	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reg.Username = strings.TrimSpace(reg.Username)
	reg.Email = strings.TrimSpace(strings.ToLower(reg.Email))
	reg.Password = strings.TrimSpace(reg.Password)

	// Validate request
	if reg.Username == "" || reg.Email == "" || reg.Password == "" {
		utils.ErrorResponse(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Username: 3–50 chars, letters/numbers/underscores only
	if !utils.UsernameRegex.MatchString(reg.Username) {
		utils.ErrorResponse(w, "Username must be 3-50 characters, letters/numbers/underscores only", http.StatusBadRequest)
		return
	}

	// Email: trim, lowercase, parse, and enforce ending in .com
	cleanEmail, err := utils.ValidateEmail(reg.Email)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	reg.Email = cleanEmail

	// Password: at least 8 chars, at least one letter and one digit
	if !utils.IsStrongPassword(reg.Password) {
		utils.ErrorResponse(w, "Password must be at least 8 characters, with at least one letter and one digit", http.StatusBadRequest)
		return
	}

	// Create user
	user, err := h.UserRepo.Create(reg)
	if err != nil {
		switch err {
		case repository.ErrEmailTaken:
			utils.ErrorResponse(w, "Email is already taken", http.StatusConflict)
		case repository.ErrUsernameTaken:
			utils.ErrorResponse(w, "Username is already taken", http.StatusConflict)
		default:
			utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create session after successful registration
	session, err := h.createUserSession(w, r, user)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, models.LoginResponse{
		User:      *user,
		SessionID: session.SessionID,
		CSRFToken: session.CSRFToken,
	}, http.StatusOK)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var login models.UserLogin
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if login.Email == "" || login.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.UserRepo.Authenticate(login)
	if err != nil {
		if err == repository.ErrInvalidCredentials {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create session after successful authentication
	session, err := h.createUserSession(w, r, user)
	if err != nil {
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
// Logout handles user logout
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
	
	// Clear the additional frontend CSRF cookie if set
    // http.SetCookie(w, &http.Cookie{
    //     Name:     "csrf_token_frontend",
    //     Value:    "",
    //     Path:     "/",
    //     MaxAge:   -1,
    //     HttpOnly: false,
    //     Secure:   false,
    //     SameSite: http.SameSiteLaxMode,
    // })

	w.WriteHeader(http.StatusOK)
}


// VerifySession handles session verification
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

// createUserSession is a helper method to create a session and set cookie
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
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, user, http.StatusOK)
}
