// Package auth contains authentication use cases and token contracts.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired access token")

// Identity represents the authenticated user information stored in a valid token.
type Identity struct {
	UserID   string
	Username string
}

// TokenIssuer creates access tokens.
type TokenIssuer interface {
	Issue(user models.User) (string, error)
}

// TokenVerifier validates an access token and extracts its identity.
type TokenVerifier interface {
	Verify(token string) (Identity, error)
}

type claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenManager issues and validates HMAC-signed JWT access tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewTokenManager creates a JWT manager with the supplied signing secret and TTL.
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

// Issue creates an access token for user.
func (m *TokenManager) Issue(user models.User) (string, error) {
	now := m.now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// Verify validates a JWT access token and extracts its identity.
func (m *TokenManager) Verify(rawToken string) (Identity, error) {
	parsed, err := jwt.ParseWithClaims(rawToken, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return Identity{}, ErrInvalidToken
	}

	tokenClaims, ok := parsed.Claims.(*claims)
	if !ok || tokenClaims.Subject == "" || tokenClaims.Username == "" {
		return Identity{}, ErrInvalidToken
	}

	return Identity{UserID: tokenClaims.Subject, Username: tokenClaims.Username}, nil
}

var (
	_ TokenIssuer   = (*TokenManager)(nil)
	_ TokenVerifier = (*TokenManager)(nil)
)
