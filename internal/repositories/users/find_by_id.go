package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
)

// FindByID returns the safe user profile for id.
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	const query = `
		SELECT id, username, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrUserNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("find user by ID: %w", err)
	}

	return user, nil
}
