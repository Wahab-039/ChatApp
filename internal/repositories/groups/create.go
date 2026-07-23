package groups

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// Create creates a new group and adds the creator as admin.
func (r *PostgresRepository) Create(ctx context.Context, name, createdBy string) (models.Group, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Group{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const createGroupQuery = `
		INSERT INTO groups (name, created_by)
		VALUES ($1, $2)
		RETURNING id, name, created_by, created_at, updated_at`

	var group models.Group
	err = tx.QueryRow(ctx, createGroupQuery, name, createdBy).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedBy,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return models.Group{}, fmt.Errorf("create group: %w", err)
	}

	const addCreatorQuery = `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, 'admin')`

	_, err = tx.Exec(ctx, addCreatorQuery, group.ID, createdBy)
	if err != nil {
		return models.Group{}, fmt.Errorf("add creator as admin: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Group{}, fmt.Errorf("commit transaction: %w", err)
	}

	return group, nil
}
