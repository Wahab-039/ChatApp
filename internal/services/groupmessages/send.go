package groupmessages

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

// Send validates, persists, and publishes a group message.
func (s *Service) Send(
	ctx context.Context,
	senderID, groupID, body, clientMessageID string,
) (SendResult, error) {
	// Validate membership
	isMember, err := s.groups.IsMember(ctx, groupID, senderID)
	if err != nil {
		return SendResult{}, err
	}
	if !isMember {
		return SendResult{}, models.ErrNotGroupMember
	}

	// Normalize input
	normalizedBody, err := normalizeBody(body)
	if err != nil {
		return SendResult{}, err
	}
	normalizedClientID, err := normalizeClientMessageID(clientMessageID)
	if err != nil {
		return SendResult{}, err
	}

	// Check for idempotent retry
	if existing, err := s.messages.FindBySenderAndClientMessageID(ctx, senderID, normalizedClientID); err == nil {
		return SendResult{Message: existing, Created: false}, nil
	} else if err != nil && !errors.Is(err, models.ErrGroupMessageNotFound) {
		return SendResult{}, err
	}

	// Persist message
	message, err := s.messages.Create(ctx, groupID, senderID, normalizedBody, normalizedClientID)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateClientMessage) {
			existing, findErr := s.messages.FindBySenderAndClientMessageID(ctx, senderID, normalizedClientID)
			if findErr == nil {
				return SendResult{Message: existing, Created: false}, nil
			}
		}
		return SendResult{}, err
	}

	// Publish to group members (best effort)
	if err := s.publishToGroupMembers(ctx, message); err != nil {
		log.Printf("mqtt publish after group message save failed: message_id=%s err=%v", message.ID, err)
	}

	return SendResult{Message: message, Created: true}, nil
}

func (s *Service) publishToGroupMembers(ctx context.Context, message models.GroupMessage) error {
	if s.publisher == nil {
		return nil
	}

	// Get all group members
	members, err := s.groups.ListMembers(ctx, message.GroupID)
	if err != nil {
		return err
	}

	// Publish to each member's inbox
	event := appmqtt.Event{
		Type:      appmqtt.EventTypeGroupMessageNew,
		RequestID: message.ClientMessageID,
		Payload: map[string]any{
			"id":                message.ID,
			"group_id":          message.GroupID,
			"sender_id":         message.SenderID,
			"body":              message.Body,
			"client_message_id": message.ClientMessageID,
			"created_at":        message.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		},
	}

	for _, member := range members {
		if member.ID == message.SenderID {
			continue // Don't publish to sender
		}
		if err := s.publisher.PublishToUserInbox(ctx, member.ID, event); err != nil {
			log.Printf("failed to publish to member %s: %v", member.ID, err)
		}
	}

	return nil
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
