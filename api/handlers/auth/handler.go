package auth

import (
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines authentication HTTP handlers.
type HandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

// Handler handles authentication-related HTTP requests.
type Handler struct {
	auth authservice.ServiceInterface
}

// NewHandler creates an authentication handler with an explicit service dependency.
func NewHandler(authService authservice.ServiceInterface) *Handler {
	return &Handler{auth: authService}
}

var _ HandlerInterface = (*Handler)(nil)
