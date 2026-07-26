// Package groupmessages contains group message use cases.
package groupmessages

import (
	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/Wahab-039/ChatApp/internal/services/groups"
)

const (
	maxBodyLength            = 4000
	maxClientMessageIDLength = 128
	defaultHistoryLimit      = 50
	maxHistoryLimit          = 100
)

// Service handles group message use cases.
type Service struct {
	groups    groups.GroupRepository
	messages  GroupMessageRepository
	publisher InboxPublisher
}

// NewService creates a group messages service.
func NewService(groups groups.GroupRepository, messages GroupMessageRepository, publisher InboxPublisher) *Service {
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
