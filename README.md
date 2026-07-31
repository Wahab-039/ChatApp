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

### Connect to receive real-time messages

After logging in, connect to EMQX via MQTT WebSocket to receive real-time message notifications:

**Step 1: Login and get credentials**
```sh
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"your-password"}'
```

Response:
```json
{
  "message": "login successful",
  "access_token": "eyJhbGc..."
}
```

**Step 2: Get your user ID**
```sh
curl http://localhost:8080/api/v1/me \
  -H 'Authorization: Bearer eyJhbGc...'
```

Response:
```json
{
  "user": {
    "id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "username": "alice",
    "created_at": "2026-07-29T10:00:00.000Z",
    "updated_at": "2026-07-29T10:00:00.000Z"
  }
}
```

**Step 3: Connect via MQTT WebSocket**

Connection parameters:
- **URL:** `ws://localhost:8083/mqtt`
- **Username:** Your `user.id` (e.g., `01933c8a-1234-7c3a-b8d5-abcdef123456`)
- **Password:** Your `access_token` (e.g., `eyJhbGc...`)
- **Client ID:** Any unique string (e.g., `alice-web-1`)

**Step 4: Subscribe to your inbox**

Topic: `chat/user/{your_user_id}/inbox`

Example: `chat/user/01933c8a-1234-7c3a-b8d5-abcdef123456/inbox`

**What you'll receive:**

Direct message event:
```json
{
  "type": "message.new",
  "request_id": "device-msg-1",
  "payload": {
    "id": "01933c8a-4f2e-7c3a-b8d5-123456789abc",
    "sender_id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
    "recipient_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "body": "hey, are you free?",
    "client_message_id": "device-msg-1",
    "created_at": "2026-07-30T12:34:56.789Z"
  }
}
```

Group message event:
```json
{
  "type": "group_message.new",
  "request_id": "device-group-msg-1",
  "payload": {
    "id": "01933c8a-8888-7c3a-b8d5-groupmsg12345",
    "group_id": "01933c8a-9999-7c3a-b8d5-group1234567",
    "sender_id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
    "body": "meeting at 3pm?",
    "client_message_id": "device-group-msg-1",
    "created_at": "2026-07-30T13:00:00.000Z"
  }
}
```

**Using MQTTX:**
1. Create New Connection
2. Name: `Alice`
3. Host: `ws://localhost:8083/mqtt`
4. Username: `01933c8a-1234-7c3a-b8d5-abcdef123456` (your user ID)
5. Password: `eyJhbGc...` (your access token)
6. Click Connect
7. In the "New Subscription" section:
   - **Topic:** `chat/user/01933c8a-1234-7c3a-b8d5-abcdef123456/inbox`
   - **QoS:** `1` (at least once delivery)
   - Click Subscribe

**Security:** EMQX ACL rules prevent users from subscribing to other users' inboxes. Attempting to subscribe to another user's inbox will be denied.

### Send a direct message

**POST** `/api/v1/messages/direct`

```sh
curl -X POST http://localhost:8080/api/v1/messages/direct \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "recipient_username": "bob",
    "body": "hey, are you free?",
    "client_message_id": "device-msg-1"
  }'
```

**Response (201 Created):**
```json
{
  "message": {
    "id": "01933c8a-4f2e-7c3a-b8d5-123456789abc",
    "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "recipient_id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
    "body": "hey, are you free?",
    "client_message_id": "device-msg-1",
    "created_at": "2026-07-30T12:34:56.789Z"
  }
}
```

Recipient (subscribed to `chat/user/{recipient_id}/inbox` via MQTT) receives:
```json
{
  "type": "message.new",
  "request_id": "device-msg-1",
  "payload": {
    "id": "01933c8a-4f2e-7c3a-b8d5-123456789abc",
    "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "recipient_id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
    "body": "hey, are you free?",
    "client_message_id": "device-msg-1",
    "created_at": "2026-07-30T12:34:56.789Z"
  }
}
```

`client_message_id` makes retries idempotent per sender. Resending the same `client_message_id` returns `200 OK` with the existing message.

### List direct message history

**GET** `/api/v1/messages/direct?with={username}&limit={limit}&before={message_id}&after={message_id}`

```sh
curl "http://localhost:8080/api/v1/messages/direct?with=bob&limit=50" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": "01933c8a-4f2e-7c3a-b8d5-123456789abc",
      "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
      "recipient_id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
      "body": "hey, are you free?",
      "client_message_id": "device-msg-1",
      "created_at": "2026-07-30T12:34:56.789Z"
    }
  ],
  "next_after": "01933c8a-4f2e-7c3a-b8d5-123456789abc",
  "next_before": "01933c8a-4f2e-7c3a-b8d5-000000000000"
}
```

**Query parameters:**
- `with` (required) — peer username
- `limit` (optional, 1-100, default 50) — messages per page
- `after` (optional) — messages newer than cursor (reconnect sync)
- `before` (optional) — older page while scrolling up

Cursors are mutually exclusive. Results are always oldest → newest.

### Create a group

**POST** `/api/v1/groups`

```sh
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Team Chat"
  }'
```

**Response (201 Created):**
```json
{
  "group": {
    "id": "01933c8a-9999-7c3a-b8d5-group1234567",
    "name": "Team Chat",
    "created_by": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "created_at": "2026-07-30T12:45:00.000Z",
    "updated_at": "2026-07-30T12:45:00.000Z"
  }
}
```

Creator is automatically added as admin.

### List my groups

**GET** `/api/v1/groups`

```sh
curl http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Response (200 OK):**
```json
{
  "groups": [
    {
      "id": "01933c8a-9999-7c3a-b8d5-group1234567",
      "name": "Team Chat",
      "created_by": "01933c8a-1234-7c3a-b8d5-abcdef123456",
      "created_at": "2026-07-30T12:45:00.000Z",
      "updated_at": "2026-07-30T12:45:00.000Z"
    }
  ]
}
```

### Get group details

**GET** `/api/v1/groups/{id}`

```sh
curl http://localhost:8080/api/v1/groups/01933c8a-9999-7c3a-b8d5-group1234567 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Response (200 OK):**
```json
{
  "id": "01933c8a-9999-7c3a-b8d5-group1234567",
  "name": "Team Chat",
  "created_by": "01933c8a-1234-7c3a-b8d5-abcdef123456",
  "created_at": "2026-07-30T12:45:00.000Z",
  "updated_at": "2026-07-30T12:45:00.000Z",
  "members": [
    {
      "id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
      "username": "alice",
      "created_at": "2026-07-29T10:00:00.000Z",
      "updated_at": "2026-07-29T10:00:00.000Z"
    },
    {
      "id": "01933c8a-5678-7c3a-b8d5-fedcba654321",
      "username": "bob",
      "created_at": "2026-07-29T11:00:00.000Z",
      "updated_at": "2026-07-29T11:00:00.000Z"
    }
  ]
}
```

### Add member to group

**POST** `/api/v1/groups/{id}/members`

```sh
curl -X POST http://localhost:8080/api/v1/groups/01933c8a-9999-7c3a-b8d5-group1234567/members \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "charlie"
  }'
```

**Response (200 OK):**
```json
{
  "message": "member added successfully"
}
```

Only existing group members can add others. New member is added with "member" role.

### Send a group message

**POST** `/api/v1/groups/{id}/messages`

```sh
curl -X POST http://localhost:8080/api/v1/groups/01933c8a-9999-7c3a-b8d5-group1234567/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "body": "meeting at 3pm?",
    "client_message_id": "device-group-msg-1"
  }'
```

**Response (201 Created):**
```json
{
  "message": {
    "id": "01933c8a-8888-7c3a-b8d5-groupmsg12345",
    "group_id": "01933c8a-9999-7c3a-b8d5-group1234567",
    "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "body": "meeting at 3pm?",
    "client_message_id": "device-group-msg-1",
    "created_at": "2026-07-30T13:00:00.000Z"
  }
}
```

All group members (except sender) receive on their inbox:
```json
{
  "type": "group_message.new",
  "request_id": "device-group-msg-1",
  "payload": {
    "id": "01933c8a-8888-7c3a-b8d5-groupmsg12345",
    "group_id": "01933c8a-9999-7c3a-b8d5-group1234567",
    "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
    "body": "meeting at 3pm?",
    "client_message_id": "device-group-msg-1",
    "created_at": "2026-07-30T13:00:00.000Z"
  }
}
```

`client_message_id` makes retries idempotent. Only group members can send messages.

### List group message history

**GET** `/api/v1/groups/{id}/messages?limit={limit}&before={message_id}&after={message_id}`

```sh
curl "http://localhost:8080/api/v1/groups/01933c8a-9999-7c3a-b8d5-group1234567/messages?limit=50" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": "01933c8a-8888-7c3a-b8d5-groupmsg12345",
      "group_id": "01933c8a-9999-7c3a-b8d5-group1234567",
      "sender_id": "01933c8a-1234-7c3a-b8d5-abcdef123456",
      "body": "meeting at 3pm?",
      "client_message_id": "device-group-msg-1",
      "created_at": "2026-07-30T13:00:00.000Z"
    }
  ],
  "next_after": "01933c8a-8888-7c3a-b8d5-groupmsg12345",
  "next_before": "01933c8a-8888-7c3a-b8d5-000000000000"
}
```

**Query parameters:**
- `limit` (optional, 1-100, default 50) — messages per page
- `after` (optional) — messages newer than cursor
- `before` (optional) — older page while scrolling up

Cursors are mutually exclusive. Only group members can view history.

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
