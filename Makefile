include .env

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATIONS_DIR=migrations

.PHONY: migrate-up migrate-down migrate-status migration

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migration:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

