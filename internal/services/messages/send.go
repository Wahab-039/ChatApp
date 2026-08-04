package messages

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

// SendDirect validates, persists, and publishes a direct message.
func (s *Service) SendDirect(
	ctx context.Context,
	senderID, recipientUsername, body string,
) (SendResult, error) {
	normalizedRecipient, err := normalizeRecipientUsername(recipientUsername)
	if err != nil {
		return SendResult{}, err
	}
	normalizedBody, err := normalizeBody(body)
	if err != nil {
		return SendResult{}, err
	}

	recipient, err := s.userRepository.FindByUsername(ctx, normalizedRecipient)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return SendResult{}, ErrRecipientNotFound
		}
		return SendResult{}, err
	}
	if recipient.ID == senderID {
		return SendResult{}, ErrCannotMessageSelf
	}

	// Generate a unique client message ID internally
	clientMessageID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), senderID[:8])

	message, err := s.messages.Create(ctx, senderID, recipient.ID, normalizedBody, clientMessageID)
	if err != nil {
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
			"sender_id":   message.SenderID,
			"body":        message.Body,
			"received_at": time.Now().UTC().Format("2006-01-02T15:04:05Z"),
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
