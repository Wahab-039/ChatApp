#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Phase 6: Group Chat ==="
echo

# Register user 1
echo "1. Registering user1..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}' | jq .

# Register user 2
echo "2. Registering user2..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"user2","password":"password123"}' | jq .

# Login user1
echo "3. Logging in as user1..."
TOKEN1=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}' | jq -r '.access_token')
echo "Token: $TOKEN1"
echo

# Login user2
echo "4. Logging in as user2..."
TOKEN2=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"user2","password":"password123"}' | jq -r '.access_token')
echo "Token: $TOKEN2"
echo

# Create a group
echo "5. Creating a group..."
GROUP_ID=$(curl -s -X POST "$BASE_URL/groups" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN1" \
  -d '{"name":"Test Group"}' | jq -r '.group.id')
echo "Group ID: $GROUP_ID"
echo

# List groups
echo "6. Listing groups for user1..."
curl -s -X GET "$BASE_URL/groups" \
  -H "Authorization: Bearer $TOKEN1" | jq .
echo

# Add user2 to group
echo "7. Adding user2 to the group..."
curl -s -X POST "$BASE_URL/groups/$GROUP_ID/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN1" \
  -d '{"username":"user2"}' | jq .
echo

# Get group details
echo "8. Getting group details..."
curl -s -X GET "$BASE_URL/groups/$GROUP_ID" \
  -H "Authorization: Bearer $TOKEN1" | jq .
echo

# Send a group message
echo "9. Sending a message to the group..."
curl -s -X POST "$BASE_URL/groups/$GROUP_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN1" \
  -d '{"body":"Hello group!","client_message_id":"msg-001"}' | jq .
echo

# Get group messages
echo "10. Getting group messages..."
curl -s -X GET "$BASE_URL/groups/$GROUP_ID/messages" \
  -H "Authorization: Bearer $TOKEN1" | jq .
echo

echo "=== Test Complete ==="
