package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func TestPostgresRepositoryIntegration(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	sqlDB, err := sql.Open("pgx", testDatabaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set Goose dialect: %v", err)
	}
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), testDatabaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	repository := NewPostgresRepository(pool)
	username := fmt.Sprintf("it_%d", time.Now().UnixNano())
	created, err := repository.Create(context.Background(), username, "test-password-hash")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", created.ID)
	})

	credentials, err := repository.FindByUsername(context.Background(), username)
	if err != nil {
		t.Fatalf("FindByUsername() error = %v", err)
	}
	if credentials.ID != created.ID || credentials.PasswordHash != "test-password-hash" {
		t.Fatalf("credentials = %#v, want created user credentials", credentials)
	}

	found, err := repository.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found != created {
		t.Fatalf("FindByID() = %#v, want %#v", found, created)
	}

	searchMatch, err := repository.Create(context.Background(), username+"_two", "test-password-hash")
	if err != nil {
		t.Fatalf("create search match: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", searchMatch.ID)
	})

	searchedUsers, err := repository.SearchByUsername(context.Background(), username, created.ID, 20)
	if err != nil {
		t.Fatalf("SearchByUsername() error = %v", err)
	}
	if len(searchedUsers) != 1 || searchedUsers[0].ID != searchMatch.ID {
		t.Fatalf("SearchByUsername() = %#v, want only %#v", searchedUsers, searchMatch)
	}

	_, err = repository.Create(context.Background(), username, "another-password-hash")
	if !errors.Is(err, models.ErrUsernameTaken) {
		t.Fatalf("duplicate Create() error = %v, want ErrUsernameTaken", err)
	}

	_, err = repository.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, models.ErrUserNotFound) {
		t.Fatalf("missing FindByID() error = %v, want ErrUserNotFound", err)
	}
}
