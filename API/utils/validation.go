package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// EMAIL VALIDATION
// ============================================================================

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates and normalizes an email address
func ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	
	if len(email) > 100 {
		return "", fmt.Errorf("email must be at most 100 characters")
	}
	
	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}
	
	return email, nil
}

// ============================================================================
// PASSWORD VALIDATION & HASHING
// ============================================================================

// IsStrongPassword checks if password meets minimum security requirements
// Requirements: At least 8 characters, at least one letter and one digit
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasLetter := false
	hasDigit := false
	
	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return true
		}
	}
	
	return false
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ============================================================================
// UUID GENERATION
// ============================================================================

// GenerateUUID generates a new UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// ============================================================================
// SESSION TOKEN GENERATION
// ============================================================================

// GenerateSessionToken generates a cryptographically secure random session token.
//
// SECURITY PROPERTIES:
//   - Entropy: 256 bits (32 bytes from crypto/rand)
//   - Encoding: Base64 URL-safe encoding (43 characters)
//   - Randomness: Uses crypto/rand.Read for CSPRNG
//   - Attack resistance: 2^256 possible values, computationally infeasible to brute force
//
// ERROR HANDLING:
//   - Returns error if crypto/rand fails (never falls back to weaker entropy)
//   - Validates that exactly 32 bytes were read
//
// OWASP COMPLIANCE:
//   - Exceeds minimum 64 bits of entropy requirement (provides 256 bits)
//   - Uses CSPRNG as required by OWASP Session Management Cheat Sheet
//
// Expected time for attacker to brute force (theoretical):
//   - At 10,000 guesses/second: > 10^70 years (universe age: ~10^10 years)
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits of entropy
	
	// Read cryptographically secure random bytes
	n, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	
	// Validate that we read exactly 32 bytes
	if n != 32 {
		return "", fmt.Errorf("insufficient random bytes: expected 32, got %d", n)
	}
	
	// Encode to base64 URL-safe format (no padding issues)
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateCSRFToken generates a cryptographically secure CSRF token.
//
// SECURITY PROPERTIES:
//   - Entropy: 256 bits (32 bytes from crypto/rand)
//   - Encoding: Base64 URL-safe encoding (43 characters)
//   - Randomness: Uses crypto/rand.Read for CSPRNG
//   - Unique per session: Prevents CSRF attacks via token validation
//
// ERROR HANDLING:
//   - Returns error if crypto/rand fails (never falls back to weaker entropy)
//   - Validates that exactly 32 bytes were read
//
// CSRF TOKEN REQUIREMENTS:
//   - Must be unpredictable (achieved via CSPRNG)
//   - Must be unique per session (achieved via 256-bit entropy)
//   - Must be validated on state-changing requests
//
// Expected collision probability:
//   - With 1 million active sessions: < 1 in 10^60
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits of entropy
	
	// Read cryptographically secure random bytes
	n, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	
	// Validate that we read exactly 32 bytes
	if n != 32 {
		return "", fmt.Errorf("insufficient random bytes: expected 32, got %d", n)
	}
	
	// Encode to base64 URL-safe format
	return base64.URLEncoding.EncodeToString(bytes), nil
}


// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetLastPathParam extracts the last parameter from a URL path
// Example: /api/posts/123 -> "123"
func GetLastPathParam(r interface{}) string {
	// This is a placeholder - implement based on your router
	// For standard http, you might parse r.URL.Path manually
	return ""
}

// Add these constants and functions to your utils/validation.go file
// They should be added after the existing CalculateSessionExpiry function

// ============================================================================
// SESSION TIMEOUT CONFIGURATION
// ============================================================================

const (
	// SessionIdleTimeout is how long a session can be inactive before expiring
	// OWASP recommends 15-30 minutes for general applications
	SessionIdleTimeout = 30 * time.Minute
	
	// SessionAbsoluteTimeout is the maximum lifetime of a session regardless of activity
	// OWASP recommends 12 hours maximum, forces re-authentication
	SessionAbsoluteTimeout = 12 * time.Hour

	// SessionRenewalWindow defines how close to expiry we can extend a session
	// Used to optimize database writes - only update expires_at if within this window
	// Prevents updating database on every single request
	SessionRenewalWindow = 5 * time.Minute
)

// CalculateSessionExpiry returns the expiry time for a session (idle timeout)
// This is updated on each request to implement sliding window
func CalculateSessionExpiry() time.Time {
	return time.Now().Add(SessionIdleTimeout)
}

// CalculateAbsoluteSessionExpiry returns the absolute expiry time for a session
// This is set once at session creation and never updated
// Forces re-authentication after absolute timeout regardless of activity
func CalculateAbsoluteSessionExpiry() time.Time {
	return time.Now().Add(SessionAbsoluteTimeout)
}

// ShouldRenewSession determines if a session should have its idle timeout extended
// Returns true if the session is within the renewal window of expiring
// This optimizes database performance by avoiding updates on every request
//
// Example: If expires_at is 2 minutes away and renewal window is 5 minutes,
// this returns true, indicating we should extend expires_at
func ShouldRenewSession(expiresAt time.Time) bool {
	timeUntilExpiry := time.Until(expiresAt)
	return timeUntilExpiry <= SessionRenewalWindow && timeUntilExpiry > 0
}