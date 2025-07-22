package middleware

import "net/http"

type CORSMiddleware struct {
	allowedOrigin string
}

func NewCORSMiddleware(origin string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigin: origin,
	}
}

func (c *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", c.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")	
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Add no-cache headers here for all responses:
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Surrogate-Control", "no-store")
		w.Header().Set("Expires", "0")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
