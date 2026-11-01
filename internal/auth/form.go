package auth

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type loginForm struct {
	Email  string
	Errors map[string]string
}

// newLoginForm initializes loginForm and sanitizes the email (lowercase + trim)
func newLoginForm(email string) *loginForm {
	return &loginForm{
		Email:  strings.ToLower(strings.TrimSpace(email)),
		Errors: make(map[string]string),
	}
}

type registerForm struct {
	Username string
	Email    string
	Errors   map[string]string
}

// newRegisterForm initializes registerForm and sanitizes username and email (lowercase + trim)
func newRegisterForm(username, email string) *registerForm {
	return &registerForm{
		Username: strings.ToLower(strings.TrimSpace(username)),
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Errors:   make(map[string]string),
	}
}

// Login form methods

func (f *loginForm) validate(password string) bool {

	// Since constructor already sanitized f.Email, we just validate it directly!
	if f.Email == "" || !emailRegex.MatchString(f.Email) {
		f.addError("email", "Please enter a valid email address.")
	}

	if password == "" {
		f.addError("password", "password is required")
	}

	return len(f.Errors) == 0
}

func (f *loginForm) addError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}

// Register methods

func (f *registerForm) validate(password string) bool {
	// Username is already lowercase & trimmed by the constructor
	if utf8.RuneCountInString(f.Username) < 3 || utf8.RuneCountInString(f.Username) > 60 {
        f.addError("username", "Username must be between 3 and 60 characters.")
	}

	// Email is already lowercase & trimmed by the constructor
	if f.Email == "" || !emailRegex.MatchString(f.Email) {
		f.addError("email", "Please enter a valid email address.")
	}

	// Passwords: ONLY trim space, NEVER lowercase!
	passwordStr := strings.TrimSpace(password)
	if utf8.RuneCountInString(passwordStr) < 3 || utf8.RuneCountInString(passwordStr) > 60 {
		f.addError("password", "Password must be between 3 and 60 characters.")
	}

	return len(f.Errors) == 0
}

func (f *registerForm) addError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}
