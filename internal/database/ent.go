package database

import (
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	entsql "entgo.io/ent/dialect/sql"
	// pgx stdlib registers the "pgx" driver name for database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewEntClient creates a new Ent client connected to PostgreSQL.
// It uses the pgx v5 stdlib driver so Ent can talk to the same
// database as your existing pgxpool, without needing a second DSN.
func NewEntClient(databaseURL string) (*ent.Client, error) {
	// Open a database/sql connection using the pgx stdlib driver.
	// entsql.Open wraps it in Ent's dialect.Driver interface.
	drv, err := entsql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open ent driver: %w", err)
	}

	// Configure the underlying *sql.DB connection pool.
	// These numbers match your existing pgxpool settings.
	sqlDB := drv.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	// Wrap the driver in the Ent client.
	// ent.Driver(drv) injects our configured driver instead of the default.
	client := ent.NewClient(ent.Driver(drv))

	return client, nil
}
