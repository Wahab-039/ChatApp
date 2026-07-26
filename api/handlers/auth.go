package handlers

import (
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	userservice "github.com/Wahab-039/ChatApp/internal/services/users"
)

// Auth handles authentication-related HTTP requests.
type Auth struct {
	auth  authservice.ServiceInterface
	users userservice.ServiceInterface
}

// NewAuth creates an authentication handler with explicit service dependencies.
func NewAuth(authService authservice.ServiceInterface, userService userservice.ServiceInterface) *Auth {
	return &Auth{auth: authService, users: userService}
}
