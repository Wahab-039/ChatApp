// Package groups contains PostgreSQL implementations of group persistence.
package groups

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository stores groups in PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a group repository backed by pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
