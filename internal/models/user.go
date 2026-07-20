// Package models contains shared application data models and model-level errors.
package models

import (
	"errors"
	"time"
)

var (
	// ErrUserNotFound is returned when no user matches a repository query.
	ErrUserNotFound = errors.New("user not found")
	// ErrUsernameTaken is returned when a username is already registered.
	ErrUsernameTaken = errors.New("username is already taken")
)

// User is the public account data persisted for a chat user.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Credentials contains the password hash used only within authentication workflows.
type Credentials struct {
	User
	PasswordHash string
}
