// Package middleware contains Gin-specific HTTP middleware.
package middleware

import (
	"net/http"
	"strings"

	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	"github.com/gin-gonic/gin"
)

const identityContextKey = "authenticated_identity"

// Auth validates bearer tokens before protected HTTP handlers run.
type Auth struct {
	tokens authservice.TokenVerifier
}

// NewAuth creates JWT middleware with an explicit token-verifier dependency.
func NewAuth(tokenVerifier authservice.TokenVerifier) *Auth {
	return &Auth{tokens: tokenVerifier}
}

// RequireAuth returns Gin middleware that requires a valid Bearer access token.
func (m *Auth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
			return
		}

		identity, err := m.tokens.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
			return
		}

		c.Set(identityContextKey, identity)
		c.Next()
	}
}

// IdentityFromContext gets the identity established by RequireAuth.
func IdentityFromContext(c *gin.Context) (authservice.Identity, bool) {
	value, ok := c.Get(identityContextKey)
	if !ok {
		return authservice.Identity{}, false
	}

	identity, ok := value.(authservice.Identity)
	return identity, ok
}
