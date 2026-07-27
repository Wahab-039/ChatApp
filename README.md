# ChatApp

A Go, Gin, and PostgreSQL backend for a real-time chat application. Authentication
and user discovery are in place. Realtime delivery uses EMQX (MQTT over
WebSocket); the Go API remains REST-first for auth and future message persistence.

## Architecture

The project follows a three-layer architecture. Dependencies always point inward:

```text
HTTP layer (api/handlers, api/middleware, api/routes)
    → service layer (internal/services)
        → model layer (internal/models)

repository implementations (internal/repositories) → model layer
internal/app assembles concrete dependencies; main.go creates and runs the app.
```

- **HTTP layer** binds requests, applies middleware, maps errors to HTTP responses,
  and returns JSON. It contains no business rules or SQL.
- **Service layer** owns use cases such as registration and login. It defines the
  repository interfaces it consumes, which keeps it independent of PostgreSQL.
- **Auth service** owns registration, login, password hashing, and JWTs.
- **User service** owns profile-related use cases, beginning with current-user lookup.
- **Repository layer** implements persistence with pgx and returns application models.
- **Model layer** holds shared business entities and model-level errors, with no Gin or
  PostgreSQL imports.
- **App package** assembles concrete infrastructure, repositories, services,
  middleware, handlers, and routes. It owns application startup and shutdown.
- **Database/config** packages are infrastructure used by `internal/app`.

## Prerequisites

- Go 1.26+
- PostgreSQL 14+
- Docker and Docker Compose (for EMQX)

## Configuration

Copy the example environment file and replace the placeholder secrets:

```sh
cp .env.example .env
```

`JWT_SECRET` must be a long, random value and must not be committed. Access-token
lifetimes use Go durations, such as `24h` or `30m`. The same secret is injected
into EMQX so MQTT clients can authenticate with the login access token.

`EMQX_SERVICE_PASSWORD` is the broker password for the Go publisher account
(`chatapp_service`). Keep it secret and out of git.

`EMQX_MQTT_TCP_URL` / `EMQX_CLIENT_ID` configure the API MQTT publisher
(Phase 2). The API connects to EMQX over MQTT TCP on startup.

## EMQX (Phase 0–2)

Start the broker:

```sh
docker compose up -d emqx
./scripts/emqx-bootstrap-service-user.sh
```

Endpoints:

| Purpose | URL |
|---|---|
| MQTT over WebSocket | `ws://localhost:8083/mqtt` |
| MQTT TCP (Go publisher) | `tcp://localhost:1883` |
| Dashboard | http://localhost:18083 |

Realtime contract (topics, ACL, JWT connect rules): [docs/realtime.md](docs/realtime.md)

Connect a client after login:

1. `POST /api/v1/auth/login` → copy `access_token`
2. `GET /api/v1/me` → copy `user.id`
3. MQTT over WebSocket:
   - username = `user.id`
   - password = `access_token`
   - subscribe = `chat/user/<user.id>/inbox`

Own inbox subscribe should succeed; another user's inbox should be denied.

### Send a direct message (Phases 3–4)

```sh
curl -sS -X POST http://localhost:8080/api/v1/messages/direct \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "recipient_username": "bob",
    "body": "hey, are you free?",
    "client_message_id": "device-msg-1"
  }'
```

Recipient MQTTX (subscribed to their inbox) should receive a `message.new` event.
`client_message_id` makes retries idempotent per sender.

### Load history / offline catch-up (Phase 5)

```sh
curl -sS "http://localhost:8080/api/v1/messages/direct?with=bob&limit=50" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Optional cursors:
- `after=<message_id>` — messages newer than last seen (reconnect sync)
- `before=<message_id>` — older page while scrolling up

### Verify Go → EMQX publish (development)

With `APP_ENV=development`, after MQTTX is subscribed to your inbox:

```sh
curl -sS -X POST http://localhost:8080/api/v1/dev/mqtt/ping \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{}'
```

MQTTX should receive a `message.new` JSON event.

## Database migrations

Create the database named in `DB_NAME`, install [Goose](https://github.com/pressly/goose),
then apply migrations (Make reads DB settings from `.env`):

```sh
make migrate-up
make migrate-status
make migrate-down
```

Create a new migration file:

```sh
make migration name=add_something
```

Or call Goose directly:

```sh
goose -dir migrations postgres "$DATABASE_URL" up
```

`DATABASE_URL` is a PostgreSQL connection string, for example:

```text
postgres://chatapp:password@localhost:5432/chatapp?sslmode=disable
```

## Run locally

```sh
go run .
```

The service listens on `http://localhost:8080` by default.

## API

### Health check

```sh
curl http://localhost:8080/health
```

The endpoint returns `503 Service Unavailable` when PostgreSQL cannot be reached.

### Register

```sh
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"wahab_039","password":"secure-password"}'
```

### Login

```sh
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"wahab_039","password":"secure-password"}'
```

Register returns a success message only. Login returns `message` and `access_token`.
Usernames must contain 3–30 lowercase letters, digits, or underscores. Passwords
must be 8–72 bytes.

### Current user

```sh
curl http://localhost:8080/api/v1/me \
  -H 'Authorization: Bearer <access_token>'
```

### Search users

```sh
curl 'http://localhost:8080/api/v1/users?query=wah' \
  -H 'Authorization: Bearer <access_token>'
```

The endpoint returns at most 20 usernames beginning with the query and excludes
the authenticated user.

## Verification

```sh
go fmt ./...
go vet ./...
go build ./...
```
