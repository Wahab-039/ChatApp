package users

import (
	userservice "github.com/Wahab-039/ChatApp/internal/services/users"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines user-profile HTTP handlers.
type HandlerInterface interface {
	Me(c *gin.Context)
	SearchUsers(c *gin.Context)
}

// Handler handles user-profile HTTP requests.
type Handler struct {
	users userservice.ServiceInterface
}

// NewHandler creates a user handler with an explicit service dependency.
func NewHandler(userService userservice.ServiceInterface) *Handler {
	return &Handler{users: userService}
}

var _ HandlerInterface = (*Handler)(nil)
