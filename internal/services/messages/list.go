package messages

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Wahab-039/ChatApp/internal/models"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 100
)

// HistoryQuery is the input for listing a direct conversation.
type HistoryQuery struct {
	PeerUsername string
	BeforeID     string
	AfterID      string
	Limit        int
}

// HistoryResult is a page of conversation messages plus optional cursors.
type HistoryResult struct {
	Messages   []models.DirectMessage
	NextBefore string
	NextAfter  string
}

// ListDirect returns a page of messages between requester and peer.
func (s *Service) ListDirect(ctx context.Context, requesterID string, query HistoryQuery) (HistoryResult, error) {
	peerUsername, err := normalizeRecipientUsername(query.PeerUsername)
	if err != nil {
		return HistoryResult{}, ErrPeerRequired
	}
	if strings.TrimSpace(query.BeforeID) != "" && strings.TrimSpace(query.AfterID) != "" {
		return HistoryResult{}, ErrInvalidCursor
	}

	limit, err := normalizeHistoryLimit(query.Limit)
	if err != nil {
		return HistoryResult{}, err
	}

	peer, err := s.users.FindByUsername(ctx, peerUsername)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return HistoryResult{}, ErrRecipientNotFound
		}
		return HistoryResult{}, err
	}
	if peer.ID == requesterID {
		return HistoryResult{}, ErrCannotMessageSelf
	}

	var before, after *models.DirectMessage
	beforeID := strings.TrimSpace(query.BeforeID)
	afterID := strings.TrimSpace(query.AfterID)

	switch {
	case beforeID != "":
		cursor, err := s.loadConversationCursor(ctx, beforeID, requesterID, peer.ID)
		if err != nil {
			return HistoryResult{}, err
		}
		before = &cursor
	case afterID != "":
		cursor, err := s.loadConversationCursor(ctx, afterID, requesterID, peer.ID)
		if err != nil {
			return HistoryResult{}, err
		}
		after = &cursor
	}

	messages, err := s.messages.ListConversation(ctx, requesterID, peer.ID, before, after, limit+1)
	if err != nil {
		return HistoryResult{}, err
	}

	hasExtra := len(messages) > limit
	if hasExtra {
		if after != nil {
			messages = messages[:limit]
		} else {
			messages = messages[1:]
		}
	}

	result := HistoryResult{Messages: messages}
	if len(result.Messages) == 0 {
		return result, nil
	}

	oldest := result.Messages[0]
	newest := result.Messages[len(result.Messages)-1]

	switch {
	case after != nil:
		result.NextBefore = oldest.ID
		if hasExtra {
			result.NextAfter = newest.ID
		}
	case before != nil:
		result.NextAfter = newest.ID
		if hasExtra {
			result.NextBefore = oldest.ID
		}
	default:
		result.NextAfter = newest.ID
		if hasExtra {
			result.NextBefore = oldest.ID
		}
	}

	return result, nil
}

func (s *Service) loadConversationCursor(
	ctx context.Context,
	messageID, requesterID, peerID string,
) (models.DirectMessage, error) {
	message, err := s.messages.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			return models.DirectMessage{}, ErrInvalidCursor
		}
		return models.DirectMessage{}, err
	}
	if !belongsToConversation(message, requesterID, peerID) {
		return models.DirectMessage{}, ErrInvalidCursor
	}
	return message, nil
}

func belongsToConversation(message models.DirectMessage, userID, peerID string) bool {
	return (message.SenderID == userID && message.RecipientID == peerID) ||
		(message.SenderID == peerID && message.RecipientID == userID)
}

func normalizeHistoryLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultHistoryLimit, nil
	}
	if limit < 1 || limit > maxHistoryLimit {
		return 0, ErrInvalidLimit
	}
	return limit, nil
}

// ParseLimit converts a query string limit into an int (0 means default).
func ParseLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, ErrInvalidLimit
	}
	return limit, nil
}
