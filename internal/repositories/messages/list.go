package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
)

// FindByID returns a direct message by id.
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (models.DirectMessage, error) {
	const query = `
		SELECT id, sender_id, recipient_id, body, client_message_id, created_at
		FROM direct_messages
		WHERE id = $1`

	var message models.DirectMessage
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		return models.DirectMessage{}, fmt.Errorf("find direct message by id: %w", err)
	}
	return message, nil
}

// ListConversation returns messages between userID and peerID.
// before/after are optional cursors; results are always oldest → newest.
func (r *PostgresRepository) ListConversation(
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

	var (
		rows pgx.Rows
		err  error
	)

	switch {
	case after != nil:
		const query = `
			SELECT id, sender_id, recipient_id, body, client_message_id, created_at
			FROM direct_messages
			WHERE (
				(sender_id = $1 AND recipient_id = $2)
				OR (sender_id = $2 AND recipient_id = $1)
			)
			AND (created_at, id) > ($3::timestamptz, $4::uuid)
			ORDER BY created_at ASC, id ASC
			LIMIT $5`
		rows, err = r.pool.Query(ctx, query, userID, peerID, after.CreatedAt.UTC(), after.ID, limit)
	case before != nil:
		const query = `
			SELECT id, sender_id, recipient_id, body, client_message_id, created_at
			FROM (
				SELECT id, sender_id, recipient_id, body, client_message_id, created_at
				FROM direct_messages
				WHERE (
					(sender_id = $1 AND recipient_id = $2)
					OR (sender_id = $2 AND recipient_id = $1)
				)
				AND (created_at, id) < ($3::timestamptz, $4::uuid)
				ORDER BY created_at DESC, id DESC
				LIMIT $5
			) page
			ORDER BY created_at ASC, id ASC`
		rows, err = r.pool.Query(ctx, query, userID, peerID, before.CreatedAt.UTC(), before.ID, limit)
	default:
		const query = `
			SELECT id, sender_id, recipient_id, body, client_message_id, created_at
			FROM (
				SELECT id, sender_id, recipient_id, body, client_message_id, created_at
				FROM direct_messages
				WHERE (
					(sender_id = $1 AND recipient_id = $2)
					OR (sender_id = $2 AND recipient_id = $1)
				)
				ORDER BY created_at DESC, id DESC
				LIMIT $3
			) page
			ORDER BY created_at ASC, id ASC`
		rows, err = r.pool.Query(ctx, query, userID, peerID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list conversation: %w", err)
	}
	defer rows.Close()

	messages := make([]models.DirectMessage, 0, limit)
	for rows.Next() {
		var message models.DirectMessage
		var createdAt time.Time
		if scanErr := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.RecipientID,
			&message.Body,
			&message.ClientMessageID,
			&createdAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scan direct message: %w", scanErr)
		}
		message.CreatedAt = createdAt
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate direct messages: %w", err)
	}

	return messages, nil
}
