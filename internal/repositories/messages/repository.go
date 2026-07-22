// Package messages contains PostgreSQL implementations of direct-message persistence.
package messages

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository stores direct messages in PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a direct-message repository backed by pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
