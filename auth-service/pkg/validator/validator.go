package validator

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Валидация email-ов
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Валидация username-ов
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,50}$`)

	// Валидация name-ов
	nameRegex = regexp.MustCompile(`^[a-zA-Z\s'-]{1,100}$`)

	// Список что бы чекать простые(распро.) пароли
	commonPasswords = map[string]bool{
		"password":    true,
		"123456":      true,
		"123456789":   true,
		"qwerty":      true,
		"12345678":    true,
		"111111":      true,
		"1234567890":  true,
		"admin":       true,
		"welcome":     true,
		"password123": true,
	}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Validator provides validation functionality
type Validator struct {
	Errors []ValidationError
}

// New creates a new validator
func New() *Validator {
	return &Validator{Errors: []ValidationError{}}
}

// Valid returns whether validation has passed
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to the validator
func (v *Validator) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Check adds an error if the condition is false
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.AddError(field, message)
	}
}

// ValidateEmail checks if the email is valid
func (v *Validator) ValidateEmail(field, email string) {
	v.Check(email != "", field, "must not be empty")

	if email == "" {
		return
	}

	v.Check(len(email) <= 255, field, "must not be more than 255 characters")
	v.Check(emailRegex.MatchString(email), field, "must be a valid email address")
}

// ValidateUsername checks if the username is valid
func (v *Validator) ValidateUsername(field, username string) {
	v.Check(username != "", field, "must not be empty")

	if username == "" {
		return
	}

	v.Check(len(username) >= 3, field, "must be at least 3 characters")
	v.Check(len(username) <= 50, field, "must not be more than 50 characters")
	v.Check(usernameRegex.MatchString(username), field, "must contain only letters, numbers, periods, underscores, or hyphens")
}

// ValidateName checks if a name is valid
func (v *Validator) ValidateName(field, name string) {
	if name == "" {
		return
	}

	v.Check(len(name) <= 100, field, "must not be more than 100 characters")
	v.Check(nameRegex.MatchString(name), field, "must contain only letters, spaces, hyphens, or apostrophes")
}

// ValidatePassword performs comprehensive password validation
func (v *Validator) ValidatePassword(field, password string) {
	v.Check(password != "", field, "must not be empty")

	if password == "" {
		return
	}

	v.Check(len(password) >= 8, field, "must be at least 8 characters")
	v.Check(len(password) <= 72, field, "must not be more than 72 characters")

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	v.Check(hasUpper, field, "must contain at least one uppercase letter")
	v.Check(hasLower, field, "must contain at least one lowercase letter")
	v.Check(hasDigit, field, "must contain at least one digit")
	v.Check(hasSpecial, field, "must contain at least one special character")

	v.Check(!commonPasswords[strings.ToLower(password)], field, "is too common, please choose a more secure password")
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() []ValidationError {
	return v.Errors
}
