// Package groupmessages contains group message use cases.
package groupmessages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	groupmessagerepository "github.com/Wahab-039/ChatApp/internal/repositories/groupmessages"
	grouprepository "github.com/Wahab-039/ChatApp/internal/repositories/groups"
)

const (
	maxBodyLength            = 4000
	maxClientMessageIDLength = 128
	defaultHistoryLimit      = 50
	maxHistoryLimit          = 100
)

// ServiceInterface defines group message use cases.
type ServiceInterface interface {
	Send(ctx context.Context, senderID, groupID, body, clientMessageID string) (SendResult, error)
	List(ctx context.Context, groupID, requesterID string, query HistoryQuery) (HistoryResult, error)
}

// Service handles group message use cases.
type Service struct {
	groups    grouprepository.RepositoryInterface
	messages  groupmessagerepository.RepositoryInterface
	publisher appmqtt.InboxPublisher
}

// NewService creates a group messages service.
func NewService(
	groups grouprepository.RepositoryInterface,
	messages groupmessagerepository.RepositoryInterface,
	publisher appmqtt.InboxPublisher,
) *Service {
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

var _ ServiceInterface = (*Service)(nil)
