// Package users contains PostgreSQL implementations of user persistence.
package users

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository stores users in PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a user repository backed by pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
