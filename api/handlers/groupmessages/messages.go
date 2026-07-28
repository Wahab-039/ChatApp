package groupmessages

import (
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	groupmessagesservice "github.com/Wahab-039/ChatApp/internal/services/groupmessages"
	"github.com/gin-gonic/gin"
)

type sendMessageRequest struct {
	Body            string `json:"body"`
	ClientMessageID string `json:"client_message_id"`
}

// SendMessage sends a message to a group.
func (h *Handler) SendMessage(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id is required"})
		return
	}

	var request sendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body and client_message_id are required"})
		return
	}

	result, err := h.messages.Send(
		c.Request.Context(),
		identity.UserID,
		groupID,
		request.Body,
		request.ClientMessageID,
	)
	if err != nil {
		h.writeMessageError(c, err, true)
		return
	}

	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{"message": result.Message})
}

// ListMessages returns paginated group message history.
func (h *Handler) ListMessages(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id is required"})
		return
	}

	limit, err := groupmessagesservice.ParseLimit(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.messages.List(c.Request.Context(), groupID, identity.UserID, groupmessagesservice.HistoryQuery{
		BeforeID: c.Query("before"),
		AfterID:  c.Query("after"),
		Limit:    limit,
	})
	if err != nil {
		h.writeMessageError(c, err, false)
		return
	}

	response := gin.H{
		"messages": result.Messages,
	}
	if result.NextBefore != "" {
		response["next_before"] = result.NextBefore
	}
	if result.NextAfter != "" {
		response["next_after"] = result.NextAfter
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) writeMessageError(c *gin.Context, err error, sending bool) {
	switch {
	case errors.Is(err, groupmessagesservice.ErrInvalidBody),
		errors.Is(err, groupmessagesservice.ErrInvalidClientMessageID),
		errors.Is(err, groupmessagesservice.ErrInvalidCursor),
		errors.Is(err, groupmessagesservice.ErrInvalidLimit):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrNotGroupMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrGroupNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case sending && errors.Is(err, models.ErrDuplicateClientMessage):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		if sending {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send message"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load messages"})
	}
}
