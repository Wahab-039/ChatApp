package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	"github.com/gin-gonic/gin"
)

type fakeAuthService struct {
	registerResult models.User
	registerErr    error
	loginResult    authservice.Result
	loginErr       error
}

func (s fakeAuthService) Register(_ context.Context, _, _ string) (models.User, error) {
	return s.registerResult, s.registerErr
}

func (s fakeAuthService) Login(_ context.Context, _, _ string) (authservice.Result, error) {
	return s.loginResult, s.loginErr
}

type fakeUserService struct {
	searchUsers []models.User
	searchErr   error
}

func (fakeUserService) CurrentUser(_ context.Context, _ string) (models.User, error) {
	return models.User{}, errors.New("not implemented")
}

func (s fakeUserService) Search(_ context.Context, _, _ string) ([]models.User, error) {
	return s.searchUsers, s.searchErr
}

func TestAuthRegisterReturnsSuccessMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", NewAuth(fakeAuthService{
		registerResult: models.User{ID: "user-1", Username: "wahab"},
	}, fakeUserService{}).Register)

	request := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"username":"wahab","password":"securepass"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if response.Body.String() != `{"message":"sign up successful"}` {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

func TestAuthLoginReturnsGenericInvalidCredentialError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", NewAuth(fakeAuthService{
		loginErr: authservice.ErrInvalidCredentials,
	}, fakeUserService{}).Login)

	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"wahab","password":"wrongpass"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if response.Body.String() != `{"error":"invalid username or password"}` {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

func TestAuthLoginReturnsOnlySuccessMessageAndToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", NewAuth(fakeAuthService{
		loginResult: authservice.Result{
			User:        models.User{ID: "user-1", Username: "wahab"},
			AccessToken: "access-token",
		},
	}, fakeUserService{}).Login)

	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"wahab","password":"securepass"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != `{"message":"login successful","access_token":"access-token"}` {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

func TestSearchUsersReturnsSafeSearchResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenManager := authservice.NewTokenManager("test-secret", time.Hour)
	token, err := tokenManager.Issue(models.User{ID: "user-1", Username: "wahab"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	router := gin.New()
	handler := NewAuth(fakeAuthService{}, fakeUserService{
		searchUsers: []models.User{{ID: "user-2", Username: "wahab_2"}},
	})
	router.GET("/users", middleware.NewAuth(tokenManager).RequireAuth(), handler.SearchUsers)

	request := httptest.NewRequest(http.MethodGet, "/users?query=wah", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Body.String() != `{"users":[{"username":"wahab_2"}]}` {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

var _ AuthService = fakeAuthService{}
var _ UserService = fakeUserService{}
