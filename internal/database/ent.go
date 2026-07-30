package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Wahab-039/ChatApp/ent"

	// pgx stdlib registers the "pgx" driver name for database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
)

const pingTimeout = 5 * time.Second

// NewEntClient creates a new Ent client and returns it along with the
// underlying *sql.DB. The *sql.DB is needed for repositories that cannot
// use Ent's query builders (e.g. tables without a surrogate id column).
func NewEntClient(databaseURL string) (*ent.Client, *sql.DB, error) {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("open sql connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	// Explicitly declare Postgres dialect so Ent generates $1/$2 placeholders
	// and double-quoted identifiers instead of MySQL-style ? and backticks.
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	client := ent.NewClient(ent.Driver(drv))

	return client, sqlDB, nil
}
