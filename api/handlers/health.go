package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Wahab-039/ChatApp/internal/database"
	"github.com/gin-gonic/gin"
)

// Health handles service health checks.
type Health struct {
	db database.HealthChecker
}

// NewHealth creates a health handler backed by database.
func NewHealth(db database.HealthChecker) *Health {
	return &Health{db: db}
}

// Check confirms the API and its database dependency are reachable.
func (h *Health) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
