package task

import (
	"strings"
	"unicode/utf8"
)

type taskForm struct {
	Title  string
	UserID *int64
	Errors map[string]string
}

func (f *taskForm) Validate() bool {
	f.Errors = make(map[string]string)

	// Validate Title
	title := strings.TrimSpace(f.Title)
	if title == "" {
		f.AddError("title", "Title is required.")
	} else if utf8.RuneCountInString(title) < 3 {
		f.AddError("title", "Title must be at least 3 characters.")
	} else if utf8.RuneCountInString(title) > 255 {
		f.AddError("title", "Title cannot exceed 255 characters.")
	}

	// Validate UserID
	if f.UserID != nil && *f.UserID <= 0 {
		f.AddError("user_id", "Invalid user selection.")
	}

	return len(f.Errors) == 0
}

func (f *taskForm) AddError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}