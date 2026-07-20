package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimiterRejectsRequestsOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewLoginRateLimiter(2, time.Minute)
	router := gin.New()
	router.POST("/login", limiter.LimitLogin(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for attempt := 1; attempt <= 3; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		request.RemoteAddr = "203.0.113.10:1234"
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		wantStatus := http.StatusNoContent
		if attempt == 3 {
			wantStatus = http.StatusTooManyRequests
		}
		if response.Code != wantStatus {
			t.Fatalf("attempt %d: status = %d, want %d", attempt, response.Code, wantStatus)
		}
	}
}

func TestLoginRateLimiterAllowsRequestsAfterWindow(t *testing.T) {
	limiter := NewLoginRateLimiter(1, time.Minute)
	currentTime := time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return currentTime }

	if !limiter.allow("203.0.113.10") {
		t.Fatal("first request was rejected")
	}
	if limiter.allow("203.0.113.10") {
		t.Fatal("second request was allowed within the window")
	}

	currentTime = currentTime.Add(time.Minute)
	if !limiter.allow("203.0.113.10") {
		t.Fatal("request was rejected after the window elapsed")
	}
}
