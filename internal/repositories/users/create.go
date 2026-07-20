package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

// Create persists a new user with an already-hashed password.
func (r *PostgresRepository) Create(ctx context.Context, username, passwordHash string) (models.User, error) {
	const query = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, created_at, updated_at`

	var user models.User
	err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, models.ErrUsernameTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
