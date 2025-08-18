#!/usr/bin/env bash
# Load variables from config.json
USERNAME=$(jq -r '.dev.username' config.json)
PASSWORD=$(jq -r '.dev.password' config.json)
PORT=$(jq -r '.dev.port' config.json)
HOST=$(jq -r '.dev.host' config.json)

echo "Logging in as ${USERNAME}..."
LOGIN_RESPONSE=$(http --form POST "${HOST}:${PORT}/login" \
  username="${USERNAME}" \
  password="${PASSWORD}" \
  Accept:application/json \
  --print=b)

# echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "ERROR: Failed to extract JWT token from login response."
  exit 1
fi

# echo "Extracted JWT token: $TOKEN"

echo "Starting game..."
http -v --json POST "${HOST}:${PORT}/admin/start" \
  Authorization:"Bearer ${TOKEN}" \
  Content-Type:application/json \
  Accept:application/json
