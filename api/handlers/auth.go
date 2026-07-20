package handlers

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
)

// AuthService defines the authentication operations used by HTTP handlers.
type AuthService interface {
	Register(ctx context.Context, username, password string) (models.User, error)
	Login(ctx context.Context, username, password string) (authservice.Result, error)
}

// UserService defines the user-profile operations used by HTTP handlers.
type UserService interface {
	CurrentUser(ctx context.Context, id string) (models.User, error)
	Search(ctx context.Context, query, requesterID string) ([]models.User, error)
}

// Auth handles authentication-related HTTP requests.
type Auth struct {
	auth  AuthService
	users UserService
}

// NewAuth creates an authentication handler with explicit service dependencies.
func NewAuth(authService AuthService, userService UserService) *Auth {
	return &Auth{auth: authService, users: userService}
}
