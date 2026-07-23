// Package groupmessages contains PostgreSQL implementations of group message persistence.
package groupmessages

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository stores group messages in PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a group messages repository backed by pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
