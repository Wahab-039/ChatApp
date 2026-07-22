# Realtime contract (EMQX + MQTT over WebSocket)

This document freezes the Phase 0 + 1 realtime contract for ChatApp.

## Transport

| Channel | URL | Used by |
|---|---|---|
| MQTT over WebSocket | `ws://localhost:8083/mqtt` | Browser / mobile clients |
| MQTT TCP | `mqtt://localhost:1883` | Future Go publisher |
| EMQX Dashboard | `http://localhost:18083` | Local broker admin |

Clients **do not** open a custom WebSocket to the Go API for chat.
They open an MQTT session to EMQX over WebSocket.

## Authentication

### End users

1. Login to the Go API: `POST /api/v1/auth/login`
2. Receive `access_token` (HS256 JWT signed with `JWT_SECRET`)
3. Connect to EMQX MQTT over WebSocket with:
   - **MQTT username** = your user UUID (`sub` claim / `GET /api/v1/me` → `user.id`)
   - **MQTT password** = the `access_token` string
   - **Client ID** = any unique value per device, for example `{userId}-{device}`

EMQX verifies the JWT with the same `JWT_SECRET` as the Go API.

### Go service account

Username: `chatapp_service` (override with `EMQX_SERVICE_USERNAME`)  
Password: value of `EMQX_SERVICE_PASSWORD` in `.env`  
Created by: `scripts/emqx-bootstrap-service-user.sh`

This account may publish to inbox/group topics. It is not an end-user login.

## Topics

| Topic | Who may subscribe | Who may publish | Purpose |
|---|---|---|---|
| `chat/user/{userId}/inbox` | That user only | `chatapp_service` only | Incoming DMs / events |
| `chat/user/{userId}/ack` | That user only | That user only | Delivery/read acks (later) |
| `chat/user/{userId}/presence` | That user only | reserved | Presence (later) |
| `chat/group/{groupId}` | group members (later ACL) | `chatapp_service` | Group fan-out (later) |

Hard rule for MVP messaging (Phase 3+):

- Clients **send** chat messages through the Go REST API.
- EMQX is used to **receive** already-persisted events.
- Clients must not publish chat payloads to another user's inbox.

## Event envelope

All broker payloads should be JSON:

```json
{
  "type": "message.new",
  "request_id": "optional-client-or-server-id",
  "payload": {}
}
```

### Initial event types

| `type` | When | Introduced |
|---|---|---|
| `message.new` | New persisted DM/group message | Phase 4 |
| `error` | Recoverable client/server realtime error | as needed |
| `message.delivered` | Delivery acknowledgement | Phase 5 |
| `message.read` | Read acknowledgement | Phase 5 |
| `presence.update` | Online/offline | Phase 7 |

### QoS

- Inbox events (`message.new`): **QoS 1** (at-least-once)
- Presence: QoS 0 is acceptable later

Clients must still treat PostgreSQL history / undelivered APIs as the source of truth after reconnect.

## ACL summary

Configured in `deploy/emqx/acl.conf`:

- `chatapp_service` → publish `chat/user/+/inbox`, `chat/group/+`
- end user `{username}` → subscribe own inbox/presence/ack
- end user `{username}` → publish own ack topic only
- everything else → deny

`authorization.no_match = deny`

## Local verification (Phase 0 + 1)

```sh
docker compose up -d emqx
./scripts/emqx-bootstrap-service-user.sh
```

Then:

1. Open dashboard: http://localhost:18083
2. Confirm listener on port 8083 / path `/mqtt`
3. Login via Go API and copy `access_token` + user id
4. In MQTTX (or similar):
   - Host: `ws://localhost:8083/mqtt`
   - Username: `<user-uuid>`
   - Password: `<access_token>`
   - Subscribe: `chat/user/<user-uuid>/inbox` → allowed
   - Subscribe: `chat/user/<other-uuid>/inbox` → denied
5. Connect with a garbage password → rejected

## Go publisher (Phase 2)

The API connects to EMQX over MQTT TCP as `chatapp_service` and publishes
JSON envelopes to user inboxes.

Development-only check (requires `APP_ENV=development` + Bearer token):

```sh
POST /api/v1/dev/mqtt/ping
```

## Direct messages (Phases 3–4)

Persist-first send:

```sh
POST /api/v1/messages/direct
```

```json
{
  "recipient_username": "bob",
  "body": "hey",
  "client_message_id": "unique-per-device-send"
}
```

After insert, the API publishes `message.new` to `chat/user/{recipientId}/inbox`.
If MQTT publish fails, the message remains in PostgreSQL.

## History / offline sync (Phase 5)

```sh
GET /api/v1/messages/direct?with=bob&limit=50
GET /api/v1/messages/direct?with=bob&after=<last_message_id>
GET /api/v1/messages/direct?with=bob&before=<oldest_loaded_id>
```

On reconnect: subscribe MQTT again, then call history with `after` to fill gaps.

## Out of scope until later

- delivered_at / read receipts
- Conversation list (“inbox of threads”)
- Groups and presence events
