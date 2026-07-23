// Package groupmessages contains group message use cases.
package groupmessages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

const (
	maxBodyLength            = 4000
	maxClientMessageIDLength = 128
	defaultHistoryLimit      = 50
	maxHistoryLimit          = 100
)

// GroupRepository checks group membership.
type GroupRepository interface {
	FindByID(ctx context.Context, id string) (models.Group, error)
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	ListMembers(ctx context.Context, groupID string) ([]models.User, error)
}

// MessageRepository persists group messages.
type MessageRepository interface {
	Create(ctx context.Context, groupID, senderID, body, clientMessageID string) (models.GroupMessage, error)
	FindBySenderAndClientMessageID(ctx context.Context, senderID, clientMessageID string) (models.GroupMessage, error)
	FindByID(ctx context.Context, id string) (models.GroupMessage, error)
	ListByGroup(ctx context.Context, groupID string, before, after *models.GroupMessage, limit int) ([]models.GroupMessage, error)
}

// InboxPublisher delivers messages to member inboxes.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event appmqtt.Event) error
}

// Service handles group message use cases.
type Service struct {
	groups    GroupRepository
	messages  MessageRepository
	publisher InboxPublisher
}

// NewService creates a group messages service.
func NewService(groups GroupRepository, messages MessageRepository, publisher InboxPublisher) *Service {
	return &Service{
		groups:    groups,
		messages:  messages,
		publisher: publisher,
	}
}

// SendResult is returned after a successful send.
type SendResult struct {
	Message models.GroupMessage
	Created bool
}
