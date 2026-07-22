// Package messages contains direct-message use cases.
package messages

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

const (
	maxBodyLength           = 4000
	maxClientMessageIDLength = 128
)

// UserRepository looks up recipients for direct messages.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (models.Credentials, error)
}

// MessageRepository persists direct messages.
type MessageRepository interface {
	Create(ctx context.Context, senderID, recipientID, body, clientMessageID string) (models.DirectMessage, error)
	FindBySenderAndClientMessageID(ctx context.Context, senderID, clientMessageID string) (models.DirectMessage, error)
	FindByID(ctx context.Context, id string) (models.DirectMessage, error)
	ListConversation(ctx context.Context, userID, peerID string, before, after *models.DirectMessage, limit int) ([]models.DirectMessage, error)
}

// InboxPublisher delivers persisted messages to recipient inboxes.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event appmqtt.Event) error
}

// Service sends and stores direct messages.
type Service struct {
	users     UserRepository
	messages  MessageRepository
	publisher InboxPublisher
}

// NewService creates a direct-message service.
func NewService(users UserRepository, messages MessageRepository, publisher InboxPublisher) *Service {
	return &Service{users: users, messages: messages, publisher: publisher}
}

// SendResult is returned after a successful send (new or idempotent replay).
type SendResult struct {
	Message models.DirectMessage
	Created bool
}

// SendDirect validates, persists, and publishes a direct message.
func (s *Service) SendDirect(
	ctx context.Context,
	senderID, recipientUsername, body, clientMessageID string,
) (SendResult, error) {
	normalizedRecipient, err := normalizeRecipientUsername(recipientUsername)
	if err != nil {
		return SendResult{}, err
	}
	normalizedBody, err := normalizeBody(body)
	if err != nil {
		return SendResult{}, err
	}
	normalizedClientID, err := normalizeClientMessageID(clientMessageID)
	if err != nil {
		return SendResult{}, err
	}

	if existing, err := s.messages.FindBySenderAndClientMessageID(ctx, senderID, normalizedClientID); err == nil {
		return SendResult{Message: existing, Created: false}, nil
	} else if err != nil && !errors.Is(err, models.ErrMessageNotFound) {
		return SendResult{}, err
	}

	recipient, err := s.users.FindByUsername(ctx, normalizedRecipient)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return SendResult{}, ErrRecipientNotFound
		}
		return SendResult{}, err
	}
	if recipient.ID == senderID {
		return SendResult{}, ErrCannotMessageSelf
	}

	message, err := s.messages.Create(ctx, senderID, recipient.ID, normalizedBody, normalizedClientID)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateClientMessage) {
			existing, findErr := s.messages.FindBySenderAndClientMessageID(ctx, senderID, normalizedClientID)
			if findErr == nil {
				return SendResult{Message: existing, Created: false}, nil
			}
		}
		return SendResult{}, err
	}

	if err := s.publishNewMessage(ctx, message); err != nil {
		log.Printf("mqtt publish after direct message save failed: message_id=%s err=%v", message.ID, err)
	}

	return SendResult{Message: message, Created: true}, nil
}

func (s *Service) publishNewMessage(ctx context.Context, message models.DirectMessage) error {
	if s.publisher == nil {
		return nil
	}
	return s.publisher.PublishToUserInbox(ctx, message.RecipientID, appmqtt.Event{
		Type:      appmqtt.EventTypeMessageNew,
		RequestID: message.ClientMessageID,
		Payload: map[string]any{
			"id":                message.ID,
			"sender_id":         message.SenderID,
			"recipient_id":      message.RecipientID,
			"body":              message.Body,
			"client_message_id": message.ClientMessageID,
			"created_at":        message.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		},
	})
}

func normalizeRecipientUsername(username string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if normalized == "" {
		return "", ErrRecipientRequired
	}
	return normalized, nil
}

func normalizeBody(body string) (string, error) {
	normalized := strings.TrimSpace(body)
	if normalized == "" {
		return "", ErrInvalidBody
	}
	if utf8.RuneCountInString(normalized) > maxBodyLength {
		return "", ErrInvalidBody
	}
	return normalized, nil
}

func normalizeClientMessageID(clientMessageID string) (string, error) {
	normalized := strings.TrimSpace(clientMessageID)
	if normalized == "" || len(normalized) > maxClientMessageIDLength {
		return "", ErrInvalidClientMessageID
	}
	return normalized, nil
}
