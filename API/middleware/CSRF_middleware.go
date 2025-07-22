package middleware

import (
	"forum/repository/session"
	"net/http"
)

func CSRFMiddleware(sessionRepo *session.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Paths to exclude from CSRF protection
			excludePaths := map[string]bool{
				"/forum/api/session/login": true,
				"/forum/api/register":      true,
			}

			// Only protect modifying methods and only if path is not excluded
			if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete) && !excludePaths[r.URL.Path] {
				cookie, err := r.Cookie("session_id")
				if err != nil || cookie.Value == "" {
					http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
					return
				}

				session, err := sessionRepo.GetBySessionID(cookie.Value)
				if err != nil {
					http.Error(w, "Invalid session", http.StatusUnauthorized)
					return
				}

				csrfHeader := r.Header.Get("X-CSRF-Token")
				if csrfHeader == "" || csrfHeader != session.CSRFToken {
					http.Error(w, "CSRF token mismatch", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
