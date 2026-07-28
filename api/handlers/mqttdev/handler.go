package mqttdev

import (
	"context"
	"net/http"
	"time"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/mqtt"
	"github.com/gin-gonic/gin"
)

// Handler handles development-only MQTT verification endpoints.
// Kept as a concrete type so routes can pass nil when disabled.
type Handler struct {
	publisher mqtt.InboxPublisher
}

// NewHandler creates a development MQTT handler.
func NewHandler(publisher mqtt.InboxPublisher) *Handler {
	return &Handler{publisher: publisher}
}

type mqttPingRequest struct {
	UserID    string         `json:"user_id"`
	RequestID string         `json:"request_id"`
	Payload   map[string]any `json:"payload"`
}

// Ping publishes a test message.new event to a user inbox (defaults to the caller).
func (h *Handler) Ping(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	var req mqttPingRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}
	}

	userID := req.UserID
	if userID == "" {
		userID = identity.UserID
	}

	payload := req.Payload
	if payload == nil {
		payload = map[string]any{
			"source":  "dev_mqtt_ping",
			"from":    identity.Username,
			"message": "hello from chatapp api",
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.publisher.PublishToUserInbox(ctx, userID, mqtt.Event{
		Type:      mqtt.EventTypeMessageNew,
		RequestID: req.RequestID,
		Payload:   payload,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to publish mqtt event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "mqtt event published",
		"topic":   mqtt.UserInboxTopic(userID),
		"user_id": userID,
	})
}
