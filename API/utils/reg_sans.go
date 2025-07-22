package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

//Only allow letters, numbers, underscores
//Length: 3–50 characters
//Prevent leading/trailing whitespace

var UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
var emailRegex = regexp.MustCompile(`^^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

//Min 8 characters
//(Optional: enforce upper/lower/digit/symbol)
//Prevent accidental whitespace

func IsStrongPassword(password string) bool {
	var hasLetter, hasDigit bool

	for _, c := range password {
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
			hasLetter = true
		case '0' <= c && c <= '9':
			hasDigit = true
		}
	}
	return len(password) >= 8 && hasLetter && hasDigit
}

// ValidateEmail trims spaces, lowercases, checks format, and enforces “.com” TLD
func ValidateEmail(raw string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(raw))

	if _, err := mail.ParseAddress(e); err != nil {
		return "", errors.New("invalid email format")
	}

	if !emailRegex.MatchString(e) {
		return "", errors.New("email must contain @ and end with .com")
	}

	return e, nil
}
