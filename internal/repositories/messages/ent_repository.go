package messages

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/ent/directmessage"
	"github.com/Wahab-039/ChatApp/internal/models"
)

// EntRepository stores direct messages using the Ent ORM client.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a messages repository backed by the Ent client.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// Compile-time check: EntRepository must satisfy RepositoryInterface.
var _ RepositoryInterface = (*EntRepository)(nil)

// toDirectMessage converts an Ent DirectMessage entity to the application model.
func toDirectMessage(e *ent.DirectMessage) models.DirectMessage {
	return models.DirectMessage{
		ID:              e.ID,
		SenderID:        e.SenderID,
		RecipientID:     e.RecipientID,
		Body:            e.Body,
		ClientMessageID: e.ClientMessageID,
		CreatedAt:       e.CreatedAt,
	}
}

// Create persists a new direct message and returns the saved entity.
func (r *EntRepository) Create(
	ctx context.Context,
	senderID, recipientID, body, clientMessageID string,
) (models.DirectMessage, error) {
	e, err := r.client.DirectMessage.Create().
		SetSenderID(senderID).
		SetRecipientID(recipientID).
		SetBody(body).
		SetClientMessageID(clientMessageID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return models.DirectMessage{}, models.ErrDuplicateClientMessage
		}
		return models.DirectMessage{}, fmt.Errorf("create direct message: %w", err)
	}

	return toDirectMessage(e), nil
}

// FindByID returns a direct message by its id.
func (r *EntRepository) FindByID(ctx context.Context, id string) (models.DirectMessage, error) {
	e, err := r.client.DirectMessage.Query().
		Where(directmessage.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.DirectMessage{}, models.ErrMessageNotFound
		}
		return models.DirectMessage{}, fmt.Errorf("find direct message by id: %w", err)
	}

	return toDirectMessage(e), nil
}

// FindBySenderAndClientMessageID looks up a message for idempotent retry handling.
func (r *EntRepository) FindBySenderAndClientMessageID(
	ctx context.Context,
	senderID, clientMessageID string,
) (models.DirectMessage, error) {
	e, err := r.client.DirectMessage.Query().
		Where(
			directmessage.SenderID(senderID),
			directmessage.ClientMessageID(clientMessageID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.DirectMessage{}, models.ErrMessageNotFound
		}
		return models.DirectMessage{}, fmt.Errorf("find direct message by client id: %w", err)
	}

	return toDirectMessage(e), nil
}

// ListConversation returns messages between userID and peerID with cursor pagination.
//
// The pgx version uses a composite (created_at, id) cursor in a single SQL comparison.
// Ent has no composite predicate, so we expand it using OR:
//
//	(created_at, id) > (T, X)
//	  ≡  created_at > T  OR  (created_at = T AND id > X)
//
// Results are always returned oldest → newest regardless of pagination direction.
// For "before" pagination we fetch DESC then reverse in memory.
func (r *EntRepository) ListConversation(
	ctx context.Context,
	userID, peerID string,
	before, after *models.DirectMessage,
	limit int,
) ([]models.DirectMessage, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if before != nil && after != nil {
		return nil, fmt.Errorf("before and after cursors are mutually exclusive")
	}

	// Base predicate: messages in either direction between the two users.
	// Matches: (sender=userID AND recipient=peerID) OR (sender=peerID AND recipient=userID)
	conversation := directmessage.Or(
		directmessage.And(
			directmessage.SenderID(userID),
			directmessage.RecipientID(peerID),
		),
		directmessage.And(
			directmessage.SenderID(peerID),
			directmessage.RecipientID(userID),
		),
	)

	switch {
	case after != nil:
		// Page forward: messages newer than the cursor, oldest→newest.
		//
		// Composite expansion: (created_at, id) > (after.CreatedAt, after.ID)
		afterTime := after.CreatedAt.UTC()
		afterID := after.ID
		cursorPredicate := directmessage.Or(
			directmessage.CreatedAtGT(afterTime),
			directmessage.And(
				directmessage.CreatedAtEQ(afterTime),
				directmessage.IDGT(afterID),
			),
		)

		entities, err := r.client.DirectMessage.Query().
			Where(conversation, cursorPredicate).
			Order(
				ent.Asc(directmessage.FieldCreatedAt),
				ent.Asc(directmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list conversation (after): %w", err)
		}
		return toDirectMessages(entities), nil

	case before != nil:
		// Page backward: messages older than the cursor.
		// Fetch DESC so LIMIT picks the closest messages, then reverse to oldest→newest.
		//
		// Composite expansion: (created_at, id) < (before.CreatedAt, before.ID)
		beforeTime := before.CreatedAt.UTC()
		beforeID := before.ID
		cursorPredicate := directmessage.Or(
			directmessage.CreatedAtLT(beforeTime),
			directmessage.And(
				directmessage.CreatedAtEQ(beforeTime),
				directmessage.IDLT(beforeID),
			),
		)

		entities, err := r.client.DirectMessage.Query().
			Where(conversation, cursorPredicate).
			Order(
				ent.Desc(directmessage.FieldCreatedAt),
				ent.Desc(directmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list conversation (before): %w", err)
		}
		// Reverse so final slice is always oldest → newest
		reverse(entities)
		return toDirectMessages(entities), nil

	default:
		// No cursor: return the most recent N messages, oldest→newest.
		// Fetch DESC so LIMIT picks the latest N, then reverse.
		entities, err := r.client.DirectMessage.Query().
			Where(conversation).
			Order(
				ent.Desc(directmessage.FieldCreatedAt),
				ent.Desc(directmessage.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("list conversation: %w", err)
		}
		reverse(entities)
		return toDirectMessages(entities), nil
	}
}

// toDirectMessages converts a slice of Ent entities to application models.
func toDirectMessages(entities []*ent.DirectMessage) []models.DirectMessage {
	result := make([]models.DirectMessage, len(entities))
	for i, e := range entities {
		result[i] = toDirectMessage(e)
	}
	return result
}

// reverse reverses a slice of DirectMessage entities in place.
func reverse(s []*ent.DirectMessage) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
