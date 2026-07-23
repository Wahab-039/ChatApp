package groupmessages

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// HistoryQuery is the input for listing group messages.
type HistoryQuery struct {
	BeforeID string
	AfterID  string
	Limit    int
}

// HistoryResult is a page of group messages plus optional cursors.
type HistoryResult struct {
	Messages   []models.GroupMessage
	NextBefore string
	NextAfter  string
}

// List returns a page of messages from a group.
func (s *Service) List(ctx context.Context, groupID, requesterID string, query HistoryQuery) (HistoryResult, error) {
	// Verify membership
	isMember, err := s.groups.IsMember(ctx, groupID, requesterID)
	if err != nil {
		return HistoryResult{}, err
	}
	if !isMember {
		return HistoryResult{}, models.ErrNotGroupMember
	}

	if strings.TrimSpace(query.BeforeID) != "" && strings.TrimSpace(query.AfterID) != "" {
		return HistoryResult{}, ErrInvalidCursor
	}

	limit, err := normalizeHistoryLimit(query.Limit)
	if err != nil {
		return HistoryResult{}, err
	}

	var before, after *models.GroupMessage
	beforeID := strings.TrimSpace(query.BeforeID)
	afterID := strings.TrimSpace(query.AfterID)

	switch {
	case beforeID != "":
		cursor, err := s.loadCursor(ctx, beforeID, groupID, requesterID)
		if err != nil {
			return HistoryResult{}, err
		}
		before = &cursor
	case afterID != "":
		cursor, err := s.loadCursor(ctx, afterID, groupID, requesterID)
		if err != nil {
			return HistoryResult{}, err
		}
		after = &cursor
	}

	messages, err := s.messages.ListByGroup(ctx, groupID, before, after, limit+1)
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

func (s *Service) loadCursor(
	ctx context.Context,
	messageID, groupID, requesterID string,
) (models.GroupMessage, error) {
	message, err := s.messages.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, models.ErrGroupMessageNotFound) {
			return models.GroupMessage{}, ErrInvalidCursor
		}
		return models.GroupMessage{}, err
	}
	if message.GroupID != groupID {
		return models.GroupMessage{}, ErrInvalidCursor
	}
	return message, nil
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
