package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	messagesservice "github.com/Wahab-039/ChatApp/internal/services/messages"
	"github.com/gin-gonic/gin"
)

// DirectMessageService defines message operations used by HTTP handlers.
type DirectMessageService interface {
	SendDirect(ctx context.Context, senderID, recipientUsername, body, clientMessageID string) (messagesservice.SendResult, error)
	ListDirect(ctx context.Context, requesterID string, query messagesservice.HistoryQuery) (messagesservice.HistoryResult, error)
}

// Messages handles direct-message HTTP requests.
type Messages struct {
	messages DirectMessageService
}

// NewMessages creates a messages handler.
func NewMessages(messages DirectMessageService) *Messages {
	return &Messages{messages: messages}
}

type sendDirectRequest struct {
	RecipientUsername string `json:"recipient_username"`
	Body              string `json:"body"`
	ClientMessageID   string `json:"client_message_id"`
}

// SendDirect persists a DM and publishes it to the recipient inbox.
func (h *Messages) SendDirect(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	var request sendDirectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_username, body, and client_message_id are required"})
		return
	}

	result, err := h.messages.SendDirect(
		c.Request.Context(),
		identity.UserID,
		request.RecipientUsername,
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

	c.JSON(status, gin.H{
		"message": result.Message,
	})
}

// ListDirect returns paginated conversation history with a peer.
func (h *Messages) ListDirect(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	limit, err := messagesservice.ParseLimit(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.messages.ListDirect(c.Request.Context(), identity.UserID, messagesservice.HistoryQuery{
		PeerUsername: c.Query("with"),
		BeforeID:     c.Query("before"),
		AfterID:      c.Query("after"),
		Limit:        limit,
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

func (h *Messages) writeMessageError(c *gin.Context, err error, sending bool) {
	switch {
	case errors.Is(err, messagesservice.ErrRecipientRequired),
		errors.Is(err, messagesservice.ErrPeerRequired),
		errors.Is(err, messagesservice.ErrInvalidBody),
		errors.Is(err, messagesservice.ErrInvalidClientMessageID),
		errors.Is(err, messagesservice.ErrCannotMessageSelf),
		errors.Is(err, messagesservice.ErrInvalidCursor),
		errors.Is(err, messagesservice.ErrInvalidLimit):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, messagesservice.ErrRecipientNotFound):
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
