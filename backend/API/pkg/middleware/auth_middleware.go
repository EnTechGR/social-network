package middleware

import (
	"context"
	"log" // Ensure log is imported
	"net/http"
	"time"

	"social-network/pkg/models"
	"social-network/pkg/repository/session"
	"social-network/pkg/repository/user_repository"
	"social-network/pkg/utils"
)

// Authentication middleware checks if the user is authenticated
type AuthMiddleware struct {
	SessionRepo *session.SessionRepository
	UserRepo    *user_repository.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(sessionRepo *session.SessionRepository, userRepo *user_repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
	}
}

// Updated Authenticate middleware with User-Agent validation
// Replace the session validation section in your Authenticate method in API/middleware/auth_middleware.go

// Authenticate middleware verifies authentication and sets user in context
// SECURITY: Implements session renewal (sliding window) to extend idle timeout for active users
// SECURITY: Validates User-Agent to detect session hijacking attempts
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(utils.GetSessionCookieName())
		if err != nil {
			// Scenario 1: No session cookie found in the request.
			log.Printf("AuthMiddleware [DEBUG]: No session cookie found for request to %s: %v", r.URL.Path, err)
			next.ServeHTTP(w, r) // Proceed as unauthenticated
			return
		}

		session, err := m.SessionRepo.GetBySessionID(cookie.Value)
		if err != nil {
			// Scenario 2: Session cookie found, but session is invalid/not found in DB.
			log.Printf("AuthMiddleware [DEBUG]: Invalid or expired session ID '%s' for request to %s: %v", cookie.Value, r.URL.Path, err)
			m.clearSessionCookie(w) // Clear potentially stale cookie
			next.ServeHTTP(w, r)    // Proceed as unauthenticated
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			// Scenario 3: Session found in DB, but its expiration time is in the past.
			log.Printf("AuthMiddleware [DEBUG]: Session ID '%s' expired (UserID: %s) for request to %s", session.SessionID, session.UserID, r.URL.Path)
			m.SessionRepo.DeleteBySessionID(session.SessionID)
			m.clearSessionCookie(w) // Clear expired cookie
			next.ServeHTTP(w, r)    // Proceed as unauthenticated
			return
		}

		// ============================================================================
		// USER-AGENT VALIDATION - NEW CODE
		// ============================================================================
		// Validate that the User-Agent matches the one stored at session creation
		// This helps detect session hijacking attempts
		
		currentUserAgent := r.Header.Get("User-Agent")
		isValid, suspicionLevel, reason := utils.ValidateUserAgent(session.UserAgent, currentUserAgent)
		
		if !isValid {
			// CRITICAL: User-Agent mismatch detected - possible session hijacking
			log.Printf("[SECURITY ALERT] Session hijacking suspected for session %s (User: %s): %s", 
				session.SessionID, session.UserID, reason)
			log.Printf("[SECURITY ALERT] Stored UA: %s", session.UserAgent)
			log.Printf("[SECURITY ALERT] Current UA: %s", currentUserAgent)
			log.Printf("[SECURITY ALERT] Request from IP: %s to %s", r.RemoteAddr, r.URL.Path)
			
			// Terminate the session immediately
			m.SessionRepo.DeleteBySessionID(session.SessionID)
			m.clearSessionCookie(w)
			
			// Return 401 Unauthorized
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "session_invalid", "message": "Session validation failed for security reasons"}`))
			return
		}
		
		// Log suspicious but allowed User-Agent variations
		if suspicionLevel != "none" {
			log.Printf("[SECURITY] User-Agent validation (Level: %s) for session %s: %s", 
				suspicionLevel, session.SessionID, reason)
		}
		// ============================================================================

		user, err := m.UserRepo.GetByID(session.UserID)
		if err != nil {
			// Scenario 4: Session is valid, but the user it points to cannot be found.
			log.Printf("AuthMiddleware [DEBUG]: User not found for session ID '%s' (UserID %s) for request to %s: %v", session.SessionID, session.UserID, r.URL.Path, err)
			m.clearSessionCookie(w) // Clear cookie, as session is invalid without a user
			next.ServeHTTP(w, r)    // Proceed as unauthenticated
			return
		}

		// Scenario 5: Authentication successful!
		log.Printf("AuthMiddleware [INFO]: User '%s' (ID: %s) authenticated for request to %s", user.Nickname, user.ID, r.URL.Path)


		// IP ADDRESS VALIDATION - NEW CODE
		// ============================================================================
		// Validate that the IP address hasn't changed drastically
		// This helps detect session hijacking from different networks
		// More flexible than User-Agent (allows same-subnet changes for DHCP/mobile)
		
		currentIP := utils.GetClientIP(r.RemoteAddr, r.Header)
		isValidIP, ipSuspicionLevel, ipReason := utils.ValidateIPAddress(session.IPAddress, currentIP)
		
		if !isValidIP {
			// CRITICAL: IP changed to different subnet - possible session hijacking
			log.Printf("[SECURITY ALERT] Suspicious IP change for session %s (User: %s): %s", 
				session.SessionID, session.UserID, ipReason)
			log.Printf("[SECURITY ALERT] Stored IP: %s", session.IPAddress)
			log.Printf("[SECURITY ALERT] Current IP: %s", currentIP)
			log.Printf("[SECURITY ALERT] Request to %s", r.URL.Path)
			
			// Analyze both IPs for detailed logging
			storedInfo := utils.AnalyzeIP(session.IPAddress)
			currentInfo := utils.AnalyzeIP(currentIP)
			log.Printf("[SECURITY ALERT] Stored subnet: %s (private: %v)", storedInfo.Subnet24, storedInfo.IsPrivate)
			log.Printf("[SECURITY ALERT] Current subnet: %s (private: %v)", currentInfo.Subnet24, currentInfo.IsPrivate)
			
			// For IP validation, we might want to be more lenient than User-Agent
			// Consider this a WARNING rather than immediate termination
			// Option 1: Terminate session (strict security)
			// Option 2: Allow but log heavily (better UX for mobile users)
			
			// STRICT MODE (uncomment to terminate session on IP mismatch):
			/*
			m.SessionRepo.DeleteBySessionID(session.SessionID)
			m.clearSessionCookie(w)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "session_invalid", "message": "Session validation failed for security reasons"}`))
			return
			*/
			
			// LENIENT MODE (current): Allow but log heavily
			// Mobile users and users behind rotating proxies will trigger this
			log.Printf("[SECURITY WARNING] Allowing request despite IP change - monitor for abuse")
		}
		
		// Log suspicious but allowed IP changes
		if ipSuspicionLevel != "none" {
			log.Printf("[SECURITY] IP validation (Level: %s) for session %s: %s (stored: %s, current: %s)", 
				ipSuspicionLevel, session.SessionID, ipReason, session.IPAddress, currentIP)
		}
		// ============================================================================

		// ============================================================================
		// SESSION RENEWAL (SLIDING WINDOW)
		// ============================================================================
		// Extend idle timeout for active users, but only if:
		// 1. Session is within renewal window (optimization - avoid DB writes on every request)
		// 2. Won't extend past absolute timeout (security - force re-auth after max time)
		
		if utils.ShouldRenewSession(session.ExpiresAt) {
			// Session is close to expiring (within renewal window)
			// Attempt to renew it to keep active user logged in
			
			newExpiresAt, renewed, err := m.SessionRepo.RenewSession(session.SessionID, session.AbsoluteExpiresAt)
			if err != nil {
				// Log but don't fail the request - session is still valid for now
				log.Printf("AuthMiddleware [WARN]: Failed to renew session %s: %v", session.SessionID, err)
			} else if renewed {
				// Session was successfully renewed
				log.Printf("[SESSION RENEWAL] Extended session %s from %v to %v (user: %s)", 
					session.SessionID, 
					session.ExpiresAt.Format("15:04:05"), 
					newExpiresAt.Format("15:04:05"),
					user.Nickname)
				
				// Update the session object in context with new expiry time
				session.ExpiresAt = newExpiresAt
			} else {
				// Renewal was skipped (would extend past absolute timeout)
				log.Printf("[SESSION RENEWAL] Skipped renewal for session %s - would exceed absolute timeout at %v", 
					session.SessionID, 
					session.AbsoluteExpiresAt.Format("15:04:05"))
			}
		}
		// ============================================================================
		
		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware ensures the user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		if user == nil {
			log.Printf("AuthMiddleware [WARN]: Authentication required for %s but user not found in context.", r.URL.Path)
			// For API requests, return JSON error
			if r.Header.Get("Accept") == "application/json" || r.Header.Get("Content-Type") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized", "message": "Authentication required"}`))
				return
			}

			// For web requests, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Printf("AuthMiddleware [DEBUG]: User found in context for authenticated request to %s.", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// RequireGuest middleware ensures the user is NOT authenticated (for login/register pages)
func (m *AuthMiddleware) RequireGuest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		if user != nil {
			log.Printf("AuthMiddleware [DEBUG]: User already authenticated for guest-only path %s. Redirecting.", r.URL.Path)
			// User is authenticated, redirect to home or dashboard
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		log.Printf("AuthMiddleware [DEBUG]: No user found in context for guest path %s. Proceeding.", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// clearSessionCookie helper function to clear session cookie
func (m *AuthMiddleware) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true, // Enable in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	log.Printf("AuthMiddleware [DEBUG]: Cleared id cookie.")
}

// GetCurrentUser returns the authenticated user from the context
func GetCurrentUser(r *http.Request) *models.User {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		// Log if user is not found or not of expected type, but avoid spamming.
		// log.Printf("GetCurrentUser: User not found in context or wrong type.")
		return nil
	}
	return user
}

// GetCurrentSession returns the current session from the context
func GetCurrentSession(r *http.Request) *models.Session {
	session, ok := r.Context().Value("session").(*models.Session)
	if !ok {
		// log.Printf("GetCurrentSession: Session not found in context or wrong type.")
		return nil
	}
	return session
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(r *http.Request) bool {
	return GetCurrentUser(r) != nil
}

// CSRF middleware to protect against CSRF attacks (especially important for OAuth)
func (m *AuthMiddleware) CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for GET, HEAD, OPTIONS requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for OAuth callback endpoints
		if r.URL.Path == "/auth/google/callback" || r.URL.Path == "/auth/github/callback" {
			log.Printf("AuthMiddleware [DEBUG]: Skipping CSRF for OAuth callback: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from form or header
		token := r.FormValue("csrf_token")
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}
		log.Printf("AuthMiddleware [DEBUG]: CSRF token received: %s (from %s)", token, r.Method)

		// Get stored CSRF token from session
		session := GetCurrentSession(r)
		if session == nil {
			log.Printf("AuthMiddleware [WARN]: CSRF check failed: No session in context for %s", r.URL.Path)
			utils.ErrorResponse(w, "Forbidden: No active session for CSRF check", http.StatusForbidden)
			return
		}

		// TODO: Implement proper CSRF token validation here
		// For now, we'll just log if a token was present (and if it matches what the session has, if it were implemented)
		if session.CSRFToken == "" {
			log.Printf("AuthMiddleware [WARN]: Session '%s' has no CSRFToken for path %s.", session.SessionID, r.URL.Path)
			// You might want to error out here in a real implementation if a session must always have a CSRF token
		} else if token == "" {
			log.Printf("AuthMiddleware [WARN]: CSRF check failed for path %s: No token provided in request.", r.URL.Path)
			utils.ErrorResponse(w, "Forbidden: CSRF token missing", http.StatusForbidden)
			return
		} else if token != session.CSRFToken {
			log.Printf("AuthMiddleware [WARN]: CSRF check failed for path %s: Mismatched token. Expected: '%s', Got: '%s'", r.URL.Path, session.CSRFToken, token)
			utils.ErrorResponse(w, "Forbidden: Invalid CSRF token", http.StatusForbidden)
			return
		}

		log.Printf("AuthMiddleware [INFO]: CSRF token valid for %s.", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
