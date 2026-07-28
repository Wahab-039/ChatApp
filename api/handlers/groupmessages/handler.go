package groupmessages

import (
	groupmessagesservice "github.com/Wahab-039/ChatApp/internal/services/groupmessages"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines group-message HTTP handlers.
type HandlerInterface interface {
	SendMessage(c *gin.Context)
	ListMessages(c *gin.Context)
}

// Handler handles group-message HTTP requests.
type Handler struct {
	messages groupmessagesservice.ServiceInterface
}

// NewHandler creates a group-messages handler.
func NewHandler(messages groupmessagesservice.ServiceInterface) *Handler {
	return &Handler{messages: messages}
}

var _ HandlerInterface = (*Handler)(nil)
