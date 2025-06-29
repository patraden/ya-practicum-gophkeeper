#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <username> <password>"
  exit 1
fi

USERNAME="$1"
PASSWORD="$2"

SERVER_HOST="localhost"
SERVER_PORT="3300"
CA_CERT="./deployments/.certs/ca.cert"
API_PATH="./api"

echo "🔐 Logging in as user: $USERNAME..."

GK_TOKEN=$(buf curl \
  --schema "$API_PATH" \
  --protocol grpc \
  --cacert "$CA_CERT" \
  --data "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
  --header "authority: $SERVER_HOST" \
  "https://$SERVER_HOST:$SERVER_PORT/gophkeeper.v1.UserService/Login" \
  | jq -r '.token')

if [[ -z "$GK_TOKEN" || "$GK_TOKEN" == "null" ]]; then
  echo "❌ Failed to retrieve token"
  exit 1
fi

echo "✅ Token acquired."
echo "$GK_TOKEN"

# 💾 Save token to <username>.token
TOKEN_FILE="${USERNAME}.token"
echo "$GK_TOKEN" > "$TOKEN_FILE"
echo "💾 Token saved to $TOKEN_FILE"