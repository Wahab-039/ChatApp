package messages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Create persists a new direct message.
func (r *PostgresRepository) Create(
	ctx context.Context,
	senderID, recipientID, body, clientMessageID string,
) (models.DirectMessage, error) {
	const query = `
		INSERT INTO direct_messages (sender_id, recipient_id, body, client_message_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, sender_id, recipient_id, body, client_message_id, created_at`

	var message models.DirectMessage
	err := r.pool.QueryRow(ctx, query, senderID, recipientID, body, clientMessageID).Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.Body,
		&message.ClientMessageID,
		&message.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.DirectMessage{}, models.ErrDuplicateClientMessage
		}
		return models.DirectMessage{}, fmt.Errorf("create direct message: %w", err)
	}

	return message, nil
}

// FindBySenderAndClientMessageID returns a previously stored message for idempotent retries.
func (r *PostgresRepository) FindBySenderAndClientMessageID(
	ctx context.Context,
	senderID, clientMessageID string,
) (models.DirectMessage, error) {
	const query = `
		SELECT id, sender_id, recipient_id, body, client_message_id, created_at
		FROM direct_messages
		WHERE sender_id = $1 AND client_message_id = $2`

	var message models.DirectMessage
	err := r.pool.QueryRow(ctx, query, senderID, clientMessageID).Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.Body,
		&message.ClientMessageID,
		&message.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.DirectMessage{}, models.ErrMessageNotFound
	}
	if err != nil {
		return models.DirectMessage{}, fmt.Errorf("find direct message by client id: %w", err)
	}

	return message, nil
}
