package auth

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Compile the regex once at package initialization.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type authForm struct {
	Username     string
	Email        string
	PasswordHash string
	Errors map[string]string
}

func (f *authForm) validate() bool {
	f.Errors = make(map[string]string)

	// Validate email
	email := strings.TrimSpace(f.Email)
	if utf8.RuneCountInString(email) < 5 {
		f.addError("email", "Must be at least 5 characters.")
	}
	
	if !emailRegex.MatchString(email) {
		f.addError("email", "Please enter a valid email address.")
	}

	// Validate password
	password := strings.TrimSpace(f.PasswordHash)
	if utf8.RuneCountInString(password) > 60 {
		f.addError("password", "Cannot exceed 60 characters.")
	}

	if utf8.RuneCountInString(password) < 5 {
		f.addError("password", "Must be at least 5 characters.")
	}

	return len(f.Errors) == 0
}

func (f *authForm) addError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}
