// Package messages contains direct-message use cases.
package messages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	messagerepository "github.com/Wahab-039/ChatApp/internal/repositories/messages"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
)

const (
	maxBodyLength            = 4000
	maxClientMessageIDLength = 128
)

// ServiceInterface defines direct-message use cases.
type ServiceInterface interface {
	SendDirect(ctx context.Context, senderID, recipientUsername, body, clientMessageID string) (SendResult, error)
	ListDirect(ctx context.Context, requesterID string, query HistoryQuery) (HistoryResult, error)
}

// Service sends and stores direct messages.
type Service struct {
	userRepository userrepository.RepositoryInterface
	messages       messagerepository.RepositoryInterface
	publisher      appmqtt.InboxPublisher
}

// NewService creates a direct-message service.
func NewService(
	userRepository userrepository.RepositoryInterface,
	messages messagerepository.RepositoryInterface,
	publisher appmqtt.InboxPublisher,
) *Service {
	return &Service{userRepository: userRepository, messages: messages, publisher: publisher}
}

// SendResult is returned after a successful send (new or idempotent replay).
type SendResult struct {
	Message models.DirectMessage
	Created bool
}

var _ ServiceInterface = (*Service)(nil)
