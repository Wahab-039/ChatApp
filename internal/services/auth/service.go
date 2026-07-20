package auth

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// UserRepository is the persistence contract required by the auth use cases.
// It is owned by the service layer so implementations do not leak into business logic.
type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (models.User, error)
	FindByUsername(ctx context.Context, username string) (models.Credentials, error)
}

// Service contains dependencies shared by authentication use cases.
type Service struct {
	users  UserRepository
	tokens TokenIssuer
}

// Result is returned after a successful login.
type Result struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"access_token"`
}

// NewService creates an authentication service with explicit dependencies.
func NewService(userRepository UserRepository, tokenIssuer TokenIssuer) *Service {
	return &Service{users: userRepository, tokens: tokenIssuer}
}
