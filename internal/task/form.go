package task

import (
	"strings"
	"unicode/utf8"
)

type taskForm struct {
	Title  string
	UserID int64
	Errors map[string]string
}

// inititalize task form
func newTaskForm(title string,userID int64) *taskForm{
	return &taskForm{
		Title: title,
		UserID: userID,
		Errors:make(map[string]string),
	}
}

func (f *taskForm) validate() bool {
	f.Errors = make(map[string]string)

	// Validate Title
	title := strings.TrimSpace(f.Title)
	if title == "" {
		f.addError("title", "Title is required.")
	} else if utf8.RuneCountInString(title) < 3 {
		f.addError("title", "Title must be at least 3 characters.")
	} else if utf8.RuneCountInString(title) > 255 {
		f.addError("title", "Title cannot exceed 255 characters.")
	}

	// Validate UserID
	if f.UserID <= 0 {
		f.addError("user_id", "Invalid user selection.")
	}

	return len(f.Errors) == 0
}

func (f *taskForm) addError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}
