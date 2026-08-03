#!/bin/sh
set -e

echo "Starting ChatApp..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

until nc -z "${DB_HOST}" "${DB_PORT}" 2>/dev/null || [ $RETRY_COUNT -eq $MAX_RETRIES ]; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "PostgreSQL is unavailable - sleeping (attempt $RETRY_COUNT/$MAX_RETRIES)"
  sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo "ERROR: PostgreSQL did not become ready in time"
  exit 1
fi

echo "PostgreSQL is ready!"

# Install goose for migrations (download latest binary)
echo "Installing goose..."
wget -qO /tmp/goose https://github.com/pressly/goose/releases/download/v3.22.1/goose_linux_x86_64
chmod +x /tmp/goose

# Construct DATABASE_URL from environment variables
export DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Run database migrations
echo "Running database migrations..."
/tmp/goose -dir /app/migrations postgres "$DATABASE_URL" up

if [ $? -eq 0 ]; then
  echo "Migrations completed successfully!"
else
  echo "ERROR: Migration failed"
  exit 1
fi

# Start the application
echo "Starting ChatApp API server..."
exec /app/ChatApp
