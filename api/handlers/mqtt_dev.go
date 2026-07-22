package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/mqtt"
	"github.com/gin-gonic/gin"
)

// InboxPublisher is the minimal MQTT contract used by the dev ping endpoint.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event mqtt.Event) error
}

// MQTTDev handles development-only MQTT verification endpoints.
type MQTTDev struct {
	publisher InboxPublisher
}

// NewMQTTDev creates a development MQTT handler.
func NewMQTTDev(publisher InboxPublisher) *MQTTDev {
	return &MQTTDev{publisher: publisher}
}

type mqttPingRequest struct {
	UserID    string         `json:"user_id"`
	RequestID string         `json:"request_id"`
	Payload   map[string]any `json:"payload"`
}

// Ping publishes a test message.new event to a user inbox (defaults to the caller).
func (h *MQTTDev) Ping(c *gin.Context) {
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
