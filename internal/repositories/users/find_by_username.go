package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5"
)

// FindByUsername returns credentials needed to authenticate username.
func (r *PostgresRepository) FindByUsername(ctx context.Context, username string) (models.Credentials, error) {
	const query = `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	var credentials models.Credentials
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&credentials.ID,
		&credentials.Username,
		&credentials.PasswordHash,
		&credentials.CreatedAt,
		&credentials.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Credentials{}, models.ErrUserNotFound
	}
	if err != nil {
		return models.Credentials{}, fmt.Errorf("find user by username: %w", err)
	}

	return credentials, nil
}
