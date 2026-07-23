package groupmessages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Create persists a new group message.
func (r *PostgresRepository) Create(
	ctx context.Context,
	groupID, senderID, body, clientMessageID string,
) (models.GroupMessage, error) {
	const query = `
		INSERT INTO group_messages (group_id, sender_id, body, client_message_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, group_id, sender_id, body, client_message_id, created_at`

	var message models.GroupMessage
	err := r.pool.QueryRow(ctx, query, groupID, senderID, body, clientMessageID).Scan(
		&message.ID,
		&message.GroupID,
		&message.SenderID,
		&message.Body,
		&message.ClientMessageID,
		&message.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.GroupMessage{}, models.ErrDuplicateClientMessage
		}
		return models.GroupMessage{}, fmt.Errorf("create group message: %w", err)
	}

	return message, nil
}

// FindBySenderAndClientMessageID returns a previously stored message for idempotent retries.
func (r *PostgresRepository) FindBySenderAndClientMessageID(
	ctx context.Context,
	senderID, clientMessageID string,
) (models.GroupMessage, error) {
	const query = `
		SELECT id, group_id, sender_id, body, client_message_id, created_at
		FROM group_messages
		WHERE sender_id = $1 AND client_message_id = $2`

	var message models.GroupMessage
	err := r.pool.QueryRow(ctx, query, senderID, clientMessageID).Scan(
		&message.ID,
		&message.GroupID,
		&message.SenderID,
		&message.Body,
		&message.ClientMessageID,
		&message.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.GroupMessage{}, models.ErrGroupMessageNotFound
	}
	if err != nil {
		return models.GroupMessage{}, fmt.Errorf("find group message by client id: %w", err)
	}

	return message, nil
}
