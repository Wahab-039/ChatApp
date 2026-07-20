package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type loginAttemptWindow struct {
	count     int
	expiresAt time.Time
}

// LoginRateLimiter limits login requests per client IP within a fixed time window.
// It is process-local and should be replaced by a shared limiter before scaling
// the API across multiple instances.
type LoginRateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	attempts map[string]loginAttemptWindow
}

// NewLoginRateLimiter creates a limiter that permits limit attempts per client IP
// during window.
func NewLoginRateLimiter(limit int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		limit:    limit,
		window:   window,
		now:      time.Now,
		attempts: make(map[string]loginAttemptWindow),
	}
}

// LimitLogin rejects requests that exceed the configured login attempt limit.
func (l *LoginRateLimiter) LimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many login attempts; try again later",
			})
			return
		}
		c.Next()
	}
}

func (l *LoginRateLimiter) allow(clientIP string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	attempt, exists := l.attempts[clientIP]
	if !exists || !now.Before(attempt.expiresAt) {
		l.attempts[clientIP] = loginAttemptWindow{
			count:     1,
			expiresAt: now.Add(l.window),
		}
		return true
	}
	if attempt.count >= l.limit {
		return false
	}

	attempt.count++
	l.attempts[clientIP] = attempt
	return true
}
