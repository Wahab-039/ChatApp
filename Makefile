GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.27.2
MIGRATIONS_DIR := migrations

.PHONY: migrate migrate-status migrate-down

# Load .env if present, then build DATABASE_URL from DB_* vars when unset.
define LOAD_DB_ENV
	set -euo pipefail; \
	if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	if [ -z "$${DATABASE_URL:-}" ]; then \
		: "$${DB_USER:?DB_USER is required (set in .env or the environment)}"; \
		: "$${DB_PASSWORD:?DB_PASSWORD is required (set in .env or the environment)}"; \
		: "$${DB_NAME:?DB_NAME is required (set in .env or the environment)}"; \
		DB_HOST="$${DB_HOST:-localhost}"; \
		DB_PORT="$${DB_PORT:-5432}"; \
		DB_SSLMODE="$${DB_SSLMODE:-disable}"; \
		DATABASE_URL="postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSLMODE}"; \
	fi;
endef

migrate: ## Apply all pending database migrations
	@$(LOAD_DB_ENV) \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up

migrate-status: ## Show migration status
	@$(LOAD_DB_ENV) \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" status

migrate-down: ## Roll back the most recent migration
	@$(LOAD_DB_ENV) \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down
