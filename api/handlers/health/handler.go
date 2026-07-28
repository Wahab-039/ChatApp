package health

import (
	"context"
	"net/http"
	"time"

	"github.com/Wahab-039/ChatApp/internal/database"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines health-check HTTP handlers.
type HandlerInterface interface {
	Check(c *gin.Context)
}

// Handler handles service health checks.
type Handler struct {
	db database.HealthChecker
}

// NewHandler creates a health handler backed by database.
func NewHandler(db database.HealthChecker) *Handler {
	return &Handler{db: db}
}

// Check confirms the API and its database dependency are reachable.
func (h *Handler) Check(c *gin.Context) {
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

var _ HandlerInterface = (*Handler)(nil)
