package auth

import "errors"

var (
	// ErrInvalidUsername is returned when a username fails format validation.
	ErrInvalidUsername = errors.New("username must be 3-30 lowercase letters, numbers, or underscores")
	// ErrInvalidPassword is returned when a password cannot safely be hashed.
	ErrInvalidPassword = errors.New("password must be between 8 and 72 bytes")
	// ErrInvalidCredentials prevents login responses from disclosing account existence.
	ErrInvalidCredentials = errors.New("invalid username or password")
)
