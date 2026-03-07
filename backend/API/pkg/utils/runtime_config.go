package utils

import (
	"os"
	"strings"
)

const (
	defaultFrontendOrigin = "http://localhost:8081"
)

// FrontendOrigin returns the configured frontend origin for CORS/WebSocket checks.
func FrontendOrigin() string {
	origin := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
	if origin == "" {
		return defaultFrontendOrigin
	}
	return origin
}

// CookieSecureEnabled controls the Secure cookie flag through env.
// Priority:
// 1) COOKIE_SECURE=true/false style values.
// 2) APP_ENV/GO_ENV/ENVIRONMENT equals "production".
func CookieSecureEnabled() bool {
	if value, ok := parseEnvBool(os.Getenv("COOKIE_SECURE")); ok {
		return value
	}

	envKeys := []string{"APP_ENV", "GO_ENV", "ENVIRONMENT"}
	for _, key := range envKeys {
		env := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if env == "production" {
			return true
		}
	}

	return false
}

func parseEnvBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}
