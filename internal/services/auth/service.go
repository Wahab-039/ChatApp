package auth

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
)

// ServiceInterface defines authentication use cases.
type ServiceInterface interface {
	Register(ctx context.Context, username, password string) (models.User, error)
	Login(ctx context.Context, username, password string) (Result, error)
}

// Service contains dependencies shared by authentication use cases.
type Service struct {
	users  userrepository.RepositoryInterface
	tokens TokenIssuer
}

// Result is returned after a successful login.
type Result struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"access_token"`
}

// NewService creates an authentication service with explicit dependencies.
func NewService(userRepository userrepository.RepositoryInterface, tokenIssuer TokenIssuer) *Service {
	return &Service{users: userRepository, tokens: tokenIssuer}
}

var _ ServiceInterface = (*Service)(nil)
