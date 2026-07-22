#!/usr/bin/env bash
# Creates the EMQX built-in database user used by the future Go MQTT publisher.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

EMQX_API_URL="${EMQX_API_URL:-http://localhost:18083}"
EMQX_DASHBOARD_USERNAME="${EMQX_DASHBOARD_USERNAME:-admin}"
EMQX_DASHBOARD_PASSWORD="${EMQX_DASHBOARD_PASSWORD:-public}"
EMQX_SERVICE_USERNAME="${EMQX_SERVICE_USERNAME:-chatapp_service}"
EMQX_SERVICE_PASSWORD="${EMQX_SERVICE_PASSWORD:-}"

if [[ -z "$EMQX_SERVICE_PASSWORD" ]]; then
  echo "EMQX_SERVICE_PASSWORD is required in .env" >&2
  exit 1
fi

echo "Waiting for EMQX API at $EMQX_API_URL ..."
TOKEN=""
for _ in $(seq 1 30); do
  LOGIN_BODY="$(curl -sS -X POST "$EMQX_API_URL/api/v5/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$EMQX_DASHBOARD_USERNAME\",\"password\":\"$EMQX_DASHBOARD_PASSWORD\"}" || true)"
  TOKEN="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("token",""))' <<<"$LOGIN_BODY" 2>/dev/null || true)"
  if [[ -n "$TOKEN" ]]; then
    break
  fi
  sleep 2
done

if [[ -z "$TOKEN" ]]; then
  echo "Unable to login to EMQX dashboard API at $EMQX_API_URL" >&2
  exit 1
fi

AUTH_ID="password_based:built_in_database"
ENCODED_AUTH_ID="$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$AUTH_ID''', safe=''))")"

# Ensure the password_based authenticator exists (env config may already create it).
AUTH_LIST_CODE="$(curl -sS -o /tmp/emqx-auth-list.json -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$EMQX_API_URL/api/v5/authentication")"
if [[ "$AUTH_LIST_CODE" != "200" ]]; then
  echo "Failed to list authentication backends (HTTP $AUTH_LIST_CODE):" >&2
  cat /tmp/emqx-auth-list.json >&2 || true
  exit 1
fi

RESPONSE="$(curl -sS -o /tmp/emqx-bootstrap-response.json -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -X POST "$EMQX_API_URL/api/v5/authentication/$ENCODED_AUTH_ID/users" \
  -d "{\"user_id\":\"$EMQX_SERVICE_USERNAME\",\"password\":\"$EMQX_SERVICE_PASSWORD\"}")"

if [[ "$RESPONSE" == "201" || "$RESPONSE" == "200" ]]; then
  echo "Created EMQX service user '$EMQX_SERVICE_USERNAME'."
  exit 0
fi

if [[ "$RESPONSE" == "409" ]]; then
  echo "EMQX service user '$EMQX_SERVICE_USERNAME' already exists; updating password."
  curl -fsS \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -X PUT "$EMQX_API_URL/api/v5/authentication/$ENCODED_AUTH_ID/users/$EMQX_SERVICE_USERNAME" \
    -d "{\"password\":\"$EMQX_SERVICE_PASSWORD\"}" >/dev/null
  echo "Updated EMQX service user password."
  exit 0
fi

echo "Failed to bootstrap EMQX service user (HTTP $RESPONSE):" >&2
cat /tmp/emqx-bootstrap-response.json >&2 || true
exit 1
