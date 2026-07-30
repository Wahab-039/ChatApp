package groupmessages

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/ent/groupmessage"
	"github.com/Wahab-039/ChatApp/internal/models"
)

// EntRepository stores group messages using the Ent ORM client.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a group messages repository backed by the Ent client.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// Compile-time check: EntRepository must satisfy RepositoryInterface.
var _ RepositoryInterface = (*EntRepository)(nil)

// toGroupMessage converts an Ent GroupMessage entity to the application model.
func toGroupMessage(e *ent.GroupMessage) models.GroupMessage {
	return models.GroupMessage{
		ID:              e.ID,
		GroupID:         e.GroupID,
		SenderID:        e.SenderID,
		Body:            e.Body,
		ClientMessageID: e.ClientMessageID,
		CreatedAt:       e.CreatedAt,
	}
}

// Create persists a new group message and returns the saved entity.
func (r *EntRepository) Create(
	ctx context.Context,
	groupID, senderID, body, clientMessageID string,
) (models.GroupMessage, error) {
	e, err := r.client.GroupMessage.Create().
		SetGroupID(groupID).
		SetSenderID(senderID).
		SetBody(body).
		SetClientMessageID(clientMessageID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return models.GroupMessage{}, models.ErrDuplicateClientMessage
		}
		return models.GroupMessage{}, fmt.Errorf("create group message: %w", err)
	}

	return toGroupMessage(e), nil
}

// FindByID returns a group message by its id.
func (r *EntRepository) FindByID(ctx context.Context, id string) (models.GroupMessage, error) {
	e, err := r.client.GroupMessage.Query().
		Where(groupmessage.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.GroupMessage{}, models.ErrGroupMessageNotFound
		}
		return models.GroupMessage{}, fmt.Errorf("find group message by id: %w", err)
	}

	return toGroupMessage(e), nil
}

// FindBySenderAndClientMessageID looks up a message for idempotent retry handling.
func (r *EntRepository) FindBySenderAndClientMessageID(
	ctx context.Context,
	senderID, clientMessageID string,
) (models.GroupMessage, error) {
	e, err := r.client.GroupMessage.Query().
		Where(
			groupmessage.SenderID(senderID),
			groupmessage.ClientMessageID(clientMessageID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.GroupMessage{}, models.ErrGroupMessageNotFound
		}
		return models.GroupMessage{}, fmt.Errorf("find group message by client id: %w", err)
	}

	return toGroupMessage(e), nil
}

// ListByGroup returns messages for a group with cursor-based pagination.
//
// Uses the same composite cursor expansion as the messages repository:
//
//	(created_at, id) > (T, X)  ≡  created_at > T  OR  (created_at = T AND id > X)
//
// Results are always returned oldest → newest.
// For "before" and default (no cursor) cases, we fetch DESC then reverse in memory.
func (r *EntRepository) ListByGroup(
	ctx context.Context,
	groupID string,
	before, after *models.GroupMessage,
	limit int,
) ([]models.GroupMessage, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if before != nil && after != nil {
		return nil, fmt.Errorf("before and after cursors are mutually exclusive")
	}

	switch {
	case after != nil:
		// Page forward: messages newer than the cursor, oldest→newest.
		afterTime := after.CreatedAt.UTC()
		afterID := after.ID

		// Composite expansion: (created_at, id) > (afterTime, afterID)
		cursorPredicate := groupmessage.Or(
			groupmessage.CreatedAtGT(afterTime),
			groupmessage.And(
				groupmessage.CreatedAtEQ(afterTime),
				groupmessage.IDGT(afterID),
			),
		)

		entities, err := r.client.GroupMessage.Query().
			Where(
				groupmessage.GroupID(groupID),
				cursorPredicate,
			).
			Order(
				ent.Asc(groupmessage.FieldCreatedAt),
				ent.Asc(groupmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list group messages (after): %w", err)
		}
		return toGroupMessages(entities), nil

	case before != nil:
		// Page backward: messages older than the cursor.
		// Fetch DESC so LIMIT takes the closest ones, then reverse to oldest→newest.
		beforeTime := before.CreatedAt.UTC()
		beforeID := before.ID

		// Composite expansion: (created_at, id) < (beforeTime, beforeID)
		cursorPredicate := groupmessage.Or(
			groupmessage.CreatedAtLT(beforeTime),
			groupmessage.And(
				groupmessage.CreatedAtEQ(beforeTime),
				groupmessage.IDLT(beforeID),
			),
		)

		entities, err := r.client.GroupMessage.Query().
			Where(
				groupmessage.GroupID(groupID),
				cursorPredicate,
			).
			Order(
				ent.Desc(groupmessage.FieldCreatedAt),
				ent.Desc(groupmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list group messages (before): %w", err)
		}
		reverseGroupMessages(entities)
		return toGroupMessages(entities), nil

	default:
		// No cursor: return the most recent N messages, oldest→newest.
		entities, err := r.client.GroupMessage.Query().
			Where(groupmessage.GroupID(groupID)).
			Order(
				ent.Desc(groupmessage.FieldCreatedAt),
				ent.Desc(groupmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list group messages: %w", err)
		}
		reverseGroupMessages(entities)
		return toGroupMessages(entities), nil
	}
}

// toGroupMessages converts a slice of Ent entities to application models.
func toGroupMessages(entities []*ent.GroupMessage) []models.GroupMessage {
	result := make([]models.GroupMessage, len(entities))
	for i, e := range entities {
		result[i] = toGroupMessage(e)
	}
	return result
}

// reverseGroupMessages reverses a slice of GroupMessage entities in place.
func reverseGroupMessages(s []*ent.GroupMessage) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
