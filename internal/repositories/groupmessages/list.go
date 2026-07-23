package groupmessages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
)

// FindByID returns a group message by ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (models.GroupMessage, error) {
	const query = `
		SELECT id, group_id, sender_id, body, client_message_id, created_at
		FROM group_messages
		WHERE id = $1`

	var message models.GroupMessage
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		return models.GroupMessage{}, fmt.Errorf("find group message by id: %w", err)
	}
	return message, nil
}

// ListByGroup returns messages for a group with cursor pagination.
func (r *PostgresRepository) ListByGroup(
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

	var (
		rows pgx.Rows
		err  error
	)

	switch {
	case after != nil:
		const query = `
			SELECT id, group_id, sender_id, body, client_message_id, created_at
			FROM group_messages
			WHERE group_id = $1
			AND (created_at, id) > ($2::timestamptz, $3::uuid)
			ORDER BY created_at ASC, id ASC
			LIMIT $4`
		rows, err = r.pool.Query(ctx, query, groupID, after.CreatedAt.UTC(), after.ID, limit)
	case before != nil:
		const query = `
			SELECT id, group_id, sender_id, body, client_message_id, created_at
			FROM (
				SELECT id, group_id, sender_id, body, client_message_id, created_at
				FROM group_messages
				WHERE group_id = $1
				AND (created_at, id) < ($2::timestamptz, $3::uuid)
				ORDER BY created_at DESC, id DESC
				LIMIT $4
			) page
			ORDER BY created_at ASC, id ASC`
		rows, err = r.pool.Query(ctx, query, groupID, before.CreatedAt.UTC(), before.ID, limit)
	default:
		const query = `
			SELECT id, group_id, sender_id, body, client_message_id, created_at
			FROM (
				SELECT id, group_id, sender_id, body, client_message_id, created_at
				FROM group_messages
				WHERE group_id = $1
				ORDER BY created_at DESC, id DESC
				LIMIT $2
			) page
			ORDER BY created_at ASC, id ASC`
		rows, err = r.pool.Query(ctx, query, groupID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list group messages: %w", err)
	}
	defer rows.Close()

	messages := make([]models.GroupMessage, 0, limit)
	for rows.Next() {
		var message models.GroupMessage
		if err := rows.Scan(
			&message.ID,
			&message.GroupID,
			&message.SenderID,
			&message.Body,
			&message.ClientMessageID,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan group message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group messages: %w", err)
	}

	return messages, nil
}
