package groups

import (
	groupsservice "github.com/Wahab-039/ChatApp/internal/services/groups"
	"github.com/gin-gonic/gin"
)

// HandlerInterface defines group-management HTTP handlers.
type HandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	AddMember(c *gin.Context)
}

// Handler handles group-management HTTP requests.
type Handler struct {
	groups groupsservice.ServiceInterface
}

// NewHandler creates a groups handler.
func NewHandler(groups groupsservice.ServiceInterface) *Handler {
	return &Handler{groups: groups}
}

var _ HandlerInterface = (*Handler)(nil)
