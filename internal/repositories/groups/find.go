package groups

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
)

// FindByID returns a group by ID.
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (models.Group, error) {
	const query = `
		SELECT id, name, created_by, created_at, updated_at
		FROM groups
		WHERE id = $1`

	var group models.Group
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedBy,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Group{}, models.ErrGroupNotFound
	}
	if err != nil {
		return models.Group{}, fmt.Errorf("find group by id: %w", err)
	}
	return group, nil
}

// ListByUserID returns all groups a user is a member of.
func (r *PostgresRepository) ListByUserID(ctx context.Context, userID string) ([]models.Group, error) {
	const query = `
		SELECT g.id, g.name, g.created_by, g.created_at, g.updated_at
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY g.updated_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list groups by user: %w", err)
	}
	defer rows.Close()

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedBy,
			&group.CreatedAt,
			&group.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate groups: %w", err)
	}
	return groups, nil
}
