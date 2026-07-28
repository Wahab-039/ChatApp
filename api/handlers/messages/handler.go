package messages

import (
	messagesservice "github.com/Wahab-039/ChatApp/internal/services/messages"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines direct-message HTTP handlers.
type HandlerInterface interface {
	SendDirect(c *gin.Context)
	ListDirect(c *gin.Context)
}

// Handler handles direct-message HTTP requests.
type Handler struct {
	messages messagesservice.ServiceInterface
}

// NewHandler creates a messages handler.
func NewHandler(messages messagesservice.ServiceInterface) *Handler {
	return &Handler{messages: messages}
}

var _ HandlerInterface = (*Handler)(nil)
