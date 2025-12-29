// Package utils provides utility functions for validation, crypto, and ID generation
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

// GenerateSessionToken generates a secure random session token
func GenerateSessionToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to UUID if random fails
		return uuid.New().String()
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// GenerateCSRFToken generates a CSRF token
func GenerateCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return uuid.New().String()
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// CalculateSessionExpiry returns the expiry time for a session (24 hours from now)
func CalculateSessionExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
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