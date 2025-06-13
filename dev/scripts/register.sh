#!/usr/bin/env bash

set -euo pipefail

SERVER_HOST="localhost"
SERVER_PORT="3200"
CA_CERT="./deployments/.certs/ca.cert"
API_PATH="./api"
USERNAME="devuser"
PASSWORD="devpassword"
ROLE="USER_ROLE_USER"

echo "Registering user '$USERNAME' with role '$ROLE'..."

RESP=$(buf curl \
  --schema "$API_PATH" \
  --protocol grpc \
  --cacert "$CA_CERT" \
  --data "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"role\":\"$ROLE\"}" \
  --header "authority: $SERVER_HOST" \
  "https://$SERVER_HOST:$SERVER_PORT/gophkeeper.v1.UserService/Register")

TOKEN=$(echo "$RESP" | jq -r '.token // empty')
USER_ID=$(echo "$RESP" | jq -r '.userId // empty')
USER_ROLE=$(echo "$RESP" | jq -r '.role // empty')

if [[ -z "$TOKEN" ]]; then
  echo "Registration failed or no token returned."
  echo "Response:"
  echo "$RESP"
  exit 1
fi

echo "Registration successful!"
echo "User ID: $USER_ID"
echo "Role: $USER_ROLE"
echo "Token: $TOKEN"