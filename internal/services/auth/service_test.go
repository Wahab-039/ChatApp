package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wahab-039/ChatApp/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	createdUsername     string
	createdPasswordHash string
	createErr           error
	credentials         models.Credentials
	findErr             error
}

func (r *fakeUserRepository) Create(_ context.Context, username, passwordHash string) (models.User, error) {
	r.createdUsername = username
	r.createdPasswordHash = passwordHash
	if r.createErr != nil {
		return models.User{}, r.createErr
	}
	return models.User{ID: "user-1", Username: username}, nil
}

func (r *fakeUserRepository) FindByUsername(_ context.Context, _ string) (models.Credentials, error) {
	if r.findErr != nil {
		return models.Credentials{}, r.findErr
	}
	return r.credentials, nil
}

type fakeTokenIssuer struct {
	token string
	err   error
}

func (i fakeTokenIssuer) Issue(_ models.User) (string, error) {
	return i.token, i.err
}

func TestServiceRegisterNormalizesUsernameAndHashesPassword(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewService(repository, fakeTokenIssuer{token: "access-token"})

	user, err := service.Register(context.Background(), "  Wahab_039  ", "securepass")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Username != "wahab_039" {
		t.Fatalf("user username = %q, want %q", user.Username, "wahab_039")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.createdPasswordHash), []byte("securepass")); err != nil {
		t.Fatalf("created password is not a bcrypt hash: %v", err)
	}
}

func TestServiceLoginRejectsUnknownUser(t *testing.T) {
	service := NewService(
		&fakeUserRepository{findErr: models.ErrUserNotFound},
		fakeTokenIssuer{token: "access-token"},
	)

	_, err := service.Login(context.Background(), "wahab", "securepass")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestTokenManagerIssuesAndVerifiesToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)

	token, err := manager.Issue(models.User{ID: "user-1", Username: "wahab"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	identity, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if identity != (Identity{UserID: "user-1", Username: "wahab"}) {
		t.Fatalf("identity = %#v", identity)
	}
}

var _ UserRepository = (*fakeUserRepository)(nil)
