package auth

import "github.com/Wahab-039/ChatApp/internal/models"

// Identity represents the authenticated user information stored in a valid token.
type Identity struct {
	UserID   string
	Username string
}

// TokenIssuer creates JWT access tokens.
type TokenIssuer interface {
	Issue(user models.User) (string, error)
}

// TokenVerifier validates a JWT access token and extracts its identity.
type TokenVerifier interface {
	Verify(token string) (Identity, error)
}
