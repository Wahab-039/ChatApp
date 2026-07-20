package auth

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register validates and creates a user.
func (s *Service) Register(ctx context.Context, username, password string) (models.User, error) {
	normalizedUsername, err := normalizeUsername(username)
	if err != nil {
		return models.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return models.User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, normalizedUsername, string(passwordHash))
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
