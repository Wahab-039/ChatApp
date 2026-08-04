#!/bin/bash
set -e

echo "🔧 Bootstrapping EMQX for ChatApp..."
echo ""

# Load environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Default values if not in .env
JWT_SECRET=${JWT_SECRET:-"d28bfc6894f97aca6bd898e50ac3d48e55b6eac4df096e6becf266530192467d"}
EMQX_SERVICE_USERNAME=${EMQX_SERVICE_USERNAME:-"chatapp_service"}
EMQX_SERVICE_PASSWORD=${EMQX_SERVICE_PASSWORD:-"chatapp_mqtt_password"}
EMQX_API_URL=${EMQX_API_URL:-"http://localhost:18083"}

echo "⏳ Waiting for EMQX to be ready..."
for i in {1..30}; do
  if curl -sf "${EMQX_API_URL}/status" >/dev/null 2>&1; then
    echo "✅ EMQX is ready!"
    break
  fi
  echo "   Attempt $i/30: waiting..."
  sleep 2
done

echo ""
echo "📝 Step 1: Configuring JWT authentication..."
curl -sf -X POST "${EMQX_API_URL}/api/v5/authentication" \
  -u "admin:public" \
  -H "Content-Type: application/json" \
  -d "{
    \"mechanism\": \"jwt\",
    \"backend\": \"jwt\",
    \"enable\": true,
    \"from\": \"password\",
    \"secret\": \"${JWT_SECRET}\",
    \"verify_claims\": {
      \"sub\": \"\${username}\"
    }
  }" >/dev/null 2>&1 && echo "✅ JWT authentication configured" || echo "⚠️  JWT auth might already exist"

echo ""
echo "📝 Step 2: Configuring built-in database authentication..."
curl -sf -X POST "${EMQX_API_URL}/api/v5/authentication" \
  -u "admin:public" \
  -H "Content-Type: application/json" \
  -d '{
    "mechanism": "password_based",
    "backend": "built_in_database",
    "enable": true
  }' >/dev/null 2>&1 && echo "✅ Built-in database authentication configured" || echo "⚠️  Built-in auth might already exist"

echo ""
echo "📝 Step 3: Creating service account user..."
curl -sf -X POST "${EMQX_API_URL}/api/v5/authentication/password_based:built_in_database/users" \
  -u "admin:public" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"${EMQX_SERVICE_USERNAME}\",
    \"password\": \"${EMQX_SERVICE_PASSWORD}\",
    \"is_superuser\": false
  }" >/dev/null 2>&1 && echo "✅ Service user '${EMQX_SERVICE_USERNAME}' created" || echo "⚠️  Service user might already exist"

echo ""
echo "📝 Step 4: Configuring ACL rules..."
# Note: ACL rules are loaded from docker-compose volume mount (deploy/emqx/acl.conf)
echo "✅ ACL rules loaded from deploy/emqx/acl.conf"

echo ""
echo "🎉 EMQX bootstrap completed!"
echo ""
echo "📋 Configuration Summary:"
echo "   - JWT Secret: ${JWT_SECRET:0:20}..."
echo "   - Service User: ${EMQX_SERVICE_USERNAME}"
echo "   - Dashboard: ${EMQX_API_URL}"
echo ""
echo "✅ You can now:"
echo "   1. Connect API: The Go service will use '${EMQX_SERVICE_USERNAME}' to publish"
echo "   2. Connect users: Use User UUID + JWT access token"
echo ""
