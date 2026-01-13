// Add this to API/utils/validation.go or create API/config/csrf_config.go

package utils

import (
	"time"
)

// ============================================================================
// CSRF TOKEN ROTATION CONFIGURATION
// ============================================================================

// OWASP Recommendation: Rotate CSRF tokens periodically and after sensitive operations
// 
// Benefits of token rotation:
// 1. Limits attack window if token is compromised
// 2. Reduces risk from token leakage
// 3. Defense-in-depth security measure
// 4. OWASP best practice

// When to rotate CSRF tokens (trigger conditions):
// 1. Authentication changes (login, logout) - MUST
// 2. Privilege escalation - MUST
// 3. Sensitive operations (password change, email change) - SHOULD
// 4. Time-based (e.g., every 30 minutes) - OPTIONAL
// 5. Every request - TOO COMPLEX for most apps

const (
	// CSRFTokenRotationInterval defines how often to rotate tokens time-based
	// Set to 0 to disable time-based rotation
	// For social networks: 30 minutes is reasonable
	CSRFTokenRotationInterval = 30 * time.Minute
	
	// CSRFTokenRotationEnabled enables/disables token rotation
	// Set to false during development if needed
	CSRFTokenRotationEnabled = true
)

// ShouldRotateCSRFToken determines if CSRF token should be rotated based on time
// Returns true if token is older than rotation interval
func ShouldRotateCSRFToken(tokenCreatedAt time.Time) bool {
	if !CSRFTokenRotationEnabled {
		return false
	}
	
	if CSRFTokenRotationInterval == 0 {
		return false // Time-based rotation disabled
	}
	
	age := time.Since(tokenCreatedAt)
	return age >= CSRFTokenRotationInterval
}

// RotationTrigger represents reasons for CSRF token rotation
type RotationTrigger string

const (
	// Must rotate (security critical)
	TriggerLogin              RotationTrigger = "login"
	TriggerLogout             RotationTrigger = "logout"
	TriggerPrivilegeChange    RotationTrigger = "privilege_change"
	
	// Should rotate (recommended)
	TriggerPasswordChange     RotationTrigger = "password_change"
	TriggerEmailChange        RotationTrigger = "email_change"
	TriggerProfileUpdate      RotationTrigger = "profile_update"
	
	// Optional rotate
	TriggerTimeBased          RotationTrigger = "time_based"
	TriggerSensitiveOperation RotationTrigger = "sensitive_operation"
)

// RotationPolicy defines which triggers should cause rotation
// Customize based on your security requirements
var RotationPolicy = map[RotationTrigger]bool{
	// Critical - always rotate
	TriggerLogin:           true,
	TriggerLogout:          true,
	TriggerPrivilegeChange: true,
	
	// Recommended - rotate for sensitive operations
	TriggerPasswordChange:  true,
	TriggerEmailChange:     true,
	TriggerProfileUpdate:   false, // Usually not needed for profile updates
	
	// Optional
	TriggerTimeBased:          true,  // Rotate every 30 minutes
	TriggerSensitiveOperation: true,  // Developer can trigger manually
}

// ShouldRotateForTrigger checks if a given trigger should cause rotation
func ShouldRotateForTrigger(trigger RotationTrigger) bool {
	if !CSRFTokenRotationEnabled {
		return false
	}
	
	shouldRotate, exists := RotationPolicy[trigger]
	if !exists {
		// Default: don't rotate for unknown triggers
		return false
	}
	
	return shouldRotate
}