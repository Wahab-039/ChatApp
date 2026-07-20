package auth

import (
	"context"
	"errors"

	"github.com/Wahab-039/ChatApp/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Login validates credentials and issues an access token on success.
func (s *Service) Login(ctx context.Context, username, password string) (Result, error) {
	normalizedUsername, err := normalizeUsername(username)
	if err != nil {
		return Result{}, ErrInvalidCredentials
	}

	credentials, err := s.users.FindByUsername(ctx, normalizedUsername)
	if errors.Is(err, models.ErrUserNotFound) {
		return Result{}, ErrInvalidCredentials
	}
	if err != nil {
		return Result{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password)); err != nil {
		return Result{}, ErrInvalidCredentials
	}

	accessToken, err := s.tokens.Issue(credentials.User)
	if err != nil {
		return Result{}, err
	}

	return Result{User: credentials.User, AccessToken: accessToken}, nil
}
