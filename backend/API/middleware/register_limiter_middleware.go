package middleware

import (
	"net"
	"net/http"
	"social-network/utils"
	"sync"
	"time"
)

var restrict = 1 * time.Minute
var coolDown = 1 * time.Second

type rateInfo struct {
	lastAttempt     time.Time
	successfulUntil time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*rateInfo
}

// NewRateLimiter initializes the IP map and cleanup job.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*rateInfo),
	}

	// Periodic cleanup
	go func() {
		for {
			time.Sleep(restrict)
			rl.cleanup()
		}
	}()

	return rl
}

// Middleware for registration rate limiting
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)
		now := time.Now()

		rl.mu.Lock()
		info, exists := rl.clients[ip]
		if !exists {
			info = &rateInfo{}
			rl.clients[ip] = info
		}

		if info.successfulUntil.After(now) {
			remaining := info.successfulUntil.Sub(now).Round(time.Second)
			rl.mu.Unlock()
			utils.ErrorResponse(w, "Too many registrations from this IP. Please wait "+remaining.String()+".", http.StatusTooManyRequests)
			return
		}

		if info.lastAttempt.Add(coolDown).After(now) {
			rl.mu.Unlock()
			utils.ErrorResponse(w, "Please wait "+coolDown.String()+" before trying again.", http.StatusTooManyRequests)
			return
		}

		// Update last attempt before calling handler
		info.lastAttempt = now
		rl.mu.Unlock()

		// Use a ResponseWriter wrapper to capture status code
		rr := &responseRecorder{ResponseWriter: w, statusCode: 200}
		next(rr, r)

		// On successful registration (HTTP 201), lock IP for 10 mins
		if rr.statusCode == http.StatusCreated {
			rl.mu.Lock()
			info.successfulUntil = time.Now().Add(restrict)
			rl.mu.Unlock()
		}
	}
}

// Get client IP from headers or remote addr
func getRealIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // fallback
	}
	return ip
}

// ResponseWriter wrapper to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// Remove expired entries
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, info := range rl.clients {
		if info.successfulUntil.Before(now.Add(-restrict)) &&
			info.lastAttempt.Before(now.Add(-restrict)) {
			delete(rl.clients, ip)
		}
	}
}
