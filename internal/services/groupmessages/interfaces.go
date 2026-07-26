package groupmessages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

// GroupMessageRepository is the persistence contract for group message operations.
type GroupMessageRepository interface {
	Create(ctx context.Context, groupID, senderID, body, clientMessageID string) (models.GroupMessage, error)
	FindByID(ctx context.Context, id string) (models.GroupMessage, error)
	FindBySenderAndClientMessageID(ctx context.Context, senderID, clientMessageID string) (models.GroupMessage, error)
	ListByGroup(ctx context.Context, groupID string, before, after *models.GroupMessage, limit int) ([]models.GroupMessage, error)
}

// InboxPublisher delivers persisted messages to recipient inboxes via MQTT.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event appmqtt.Event) error
}
