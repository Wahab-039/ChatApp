package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	messagesservice "github.com/Wahab-039/ChatApp/internal/services/messages"
	"github.com/gin-gonic/gin"
)

type fakeMessageService struct {
	result     messagesservice.SendResult
	err        error
	listResult messagesservice.HistoryResult
	listErr    error
}

func (f fakeMessageService) SendDirect(_ context.Context, _, _, _, _ string) (messagesservice.SendResult, error) {
	return f.result, f.err
}

func (f fakeMessageService) ListDirect(_ context.Context, _ string, _ messagesservice.HistoryQuery) (messagesservice.HistoryResult, error) {
	return f.listResult, f.listErr
}

func TestSendDirectReturnsCreatedMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenManager := authservice.NewTokenManager("test-secret", time.Hour)
	token, err := tokenManager.Issue(models.User{ID: "user-a", Username: "alice"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	router := gin.New()
	router.POST("/messages/direct", middleware.NewAuth(tokenManager).RequireAuth(), NewMessages(fakeMessageService{
		result: messagesservice.SendResult{
			Created: true,
			Message: models.DirectMessage{
				ID:              "msg-1",
				SenderID:        "user-a",
				RecipientID:     "user-b",
				Body:            "hello",
				ClientMessageID: "client-1",
				CreatedAt:       time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC),
			},
		},
	}).SendDirect)

	request := httptest.NewRequest(http.MethodPost, "/messages/direct", strings.NewReader(
		`{"recipient_username":"bob","body":"hello","client_message_id":"client-1"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}
}

func TestSendDirectMapsRecipientNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenManager := authservice.NewTokenManager("test-secret", time.Hour)
	token, err := tokenManager.Issue(models.User{ID: "user-a", Username: "alice"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	router := gin.New()
	router.POST("/messages/direct", middleware.NewAuth(tokenManager).RequireAuth(), NewMessages(fakeMessageService{
		err: messagesservice.ErrRecipientNotFound,
	}).SendDirect)

	request := httptest.NewRequest(http.MethodPost, "/messages/direct", strings.NewReader(
		`{"recipient_username":"missing","body":"hello","client_message_id":"client-1"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestListDirectReturnsMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenManager := authservice.NewTokenManager("test-secret", time.Hour)
	token, err := tokenManager.Issue(models.User{ID: "user-a", Username: "alice"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	router := gin.New()
	router.GET("/messages/direct", middleware.NewAuth(tokenManager).RequireAuth(), NewMessages(fakeMessageService{
		listResult: messagesservice.HistoryResult{
			Messages: []models.DirectMessage{{
				ID:       "msg-1",
				SenderID: "user-a",
				Body:     "hello",
			}},
			NextAfter: "msg-1",
		},
	}).ListDirect)

	request := httptest.NewRequest(http.MethodGet, "/messages/direct?with=bob", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"id":"msg-1"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}
