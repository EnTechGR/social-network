package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a bcrypt hashed password with a plain password
func CheckPasswordHash(password, hash string) bool {
	fmt.Println("Checking password:", password)
	fmt.Println("Against hash:", hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}


