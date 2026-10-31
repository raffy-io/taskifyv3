package utils

import (
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain-text password and returns its bcrypt hash.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a plain-text password against a hashed password.
// Returns true if they match, false otherwise.
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		slog.Info("Password check failed!", "error", err)
	} else {
		slog.Info("Password matches perfectly!")
	}
	return err == nil
}
