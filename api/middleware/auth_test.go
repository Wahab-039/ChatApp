package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	"github.com/gin-gonic/gin"
)

type fakeTokenVerifier struct {
	identity authservice.Identity
	err      error
}

func (v fakeTokenVerifier) Verify(_ string) (authservice.Identity, error) {
	return v.identity, v.err
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		verifier   fakeTokenVerifier
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "missing bearer token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			header:     "Bearer invalid-token",
			verifier:   fakeTokenVerifier{err: errors.New("invalid token")},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token",
			header:     "Bearer valid-token",
			verifier:   fakeTokenVerifier{identity: authservice.Identity{UserID: "user-1", Username: "wahab"}},
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			nextCalled := false
			router.GET("/protected", NewAuth(test.verifier).RequireAuth(), func(c *gin.Context) {
				nextCalled = true
				if _, ok := IdentityFromContext(c); !ok {
					t.Error("IdentityFromContext() returned false")
				}
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.header != "" {
				request.Header.Set("Authorization", test.header)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if nextCalled != test.wantNext {
				t.Fatalf("next handler called = %t, want %t", nextCalled, test.wantNext)
			}
		})
	}
}

var _ authservice.TokenVerifier = fakeTokenVerifier{}
