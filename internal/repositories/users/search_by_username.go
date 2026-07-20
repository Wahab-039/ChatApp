package users

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// SearchByUsername returns users whose usernames begin with query, excluding excludedUserID.
func (r *PostgresRepository) SearchByUsername(
	ctx context.Context,
	query, excludedUserID string,
	limit int,
) ([]models.User, error) {
	const sql = `
		SELECT id, username, created_at, updated_at
		FROM users
		WHERE username LIKE $1 || '%'
		  AND id <> $2
		ORDER BY username ASC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, sql, query, excludedUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("search users by username: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan searched user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched users: %w", err)
	}

	return users, nil
}
