package groups

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

// AddMember adds a user to a group.
func (r *PostgresRepository) AddMember(ctx context.Context, groupID, userID, role string) error {
	const query = `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, groupID, userID, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrAlreadyGroupMember
		}
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

// IsMember checks if a user is a member of a group.
func (r *PostgresRepository) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE group_id = $1 AND user_id = $2
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check group membership: %w", err)
	}
	return exists, nil
}

// ListMembers returns all members of a group.
func (r *PostgresRepository) ListMembers(ctx context.Context, groupID string) ([]models.User, error) {
	const query = `
		SELECT u.id, u.username, u.created_at, u.updated_at
		FROM users u
		INNER JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	defer rows.Close()

	members := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}
	return members, nil
}
