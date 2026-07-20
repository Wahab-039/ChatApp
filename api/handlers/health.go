package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabaseHealthChecker is the minimal database contract required by the health endpoint.
type DatabaseHealthChecker interface {
	Ping(ctx context.Context) error
}

// Health handles service health checks.
type Health struct {
	database DatabaseHealthChecker
}

// NewHealth creates a health handler backed by database.
func NewHealth(database DatabaseHealthChecker) *Health {
	return &Health{database: database}
}

// Check confirms the API and its database dependency are reachable.
func (h *Health) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
	defer cancel()

	if err := h.database.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
