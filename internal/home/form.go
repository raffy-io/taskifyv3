package home

import (
	"strings"
	"unicode/utf8"
)

type editHomepageForm struct {
	Title        string
	Subtitle     string
	HomeImage    string
	HomeImageAlt string
	Errors       map[string]string
}

// NewEditHomepageForm initializes EditHomepageForm and trims whitespace from string inputs.
func newEditHomepageForm(title, subtitle, homeImage, homeImageAlt string) *editHomepageForm {
	return &editHomepageForm{
		Title:        strings.TrimSpace(title),
		Subtitle:     strings.TrimSpace(subtitle),
		HomeImage:    strings.TrimSpace(homeImage),
		HomeImageAlt: strings.TrimSpace(homeImageAlt),
		Errors:       make(map[string]string),
	}
}

// edit home form methods

func (f *editHomepageForm) validate() bool {

	if utf8.RuneCountInString(f.Title) < 3 || utf8.RuneCountInString(f.Title) > 60 {
		f.addError("title", "Title must be between 3 and 60 characters.")
	}

	if utf8.RuneCountInString(f.Subtitle) < 3 || utf8.RuneCountInString(f.Subtitle) > 60 {
		f.addError("subtitle", "Subtitle must be between 3 and 60 characters.")
	}

	if f.HomeImageAlt == "" {
		f.addError("imageAlt", "Image alt required.Important for SEO")
	}

	return len(f.Errors) == 0
}

func (f *editHomepageForm) addError(field string, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}
