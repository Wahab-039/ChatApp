package auth

import (
	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/Wahab-039/ChatApp/internal/services/users"
)

// Service contains dependencies shared by authentication use cases.
type Service struct {
	users  users.UserRepository
	tokens TokenIssuer
}

// Result is returned after a successful login.
type Result struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"access_token"`
}

// NewService creates an authentication service with explicit dependencies.
func NewService(userRepository users.UserRepository, tokenIssuer TokenIssuer) *Service {
	return &Service{users: userRepository, tokens: tokenIssuer}
}
