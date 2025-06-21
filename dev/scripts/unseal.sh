#!/usr/bin/env bash

set -euo pipefail

SERVER_HOST="localhost"
SERVER_PORT="3300"
CA_CERT="./deployments/.certs/ca.cert"
SHARES_PATH="./deployments/.crypto/shares.json"
API_PATH="./api"

echo "🔐 Logging in as default admin..."
GK_TOKEN=$(buf curl \
  --schema "$API_PATH" \
  --protocol grpc \
  --cacert "$CA_CERT" \
  --data '{"username":"Admin","password":"Admin"}' \
  --header "authority: $SERVER_HOST" \
  "https://$SERVER_HOST:$SERVER_PORT/gophkeeper.v1.UserService/Login" \
  | jq -r '.token')

if [[ -z "$GK_TOKEN" || "$GK_TOKEN" == "null" ]]; then
  echo "Failed to retrieve token"
  exit 1
fi

echo "✅ Token acquired."

echo "Submitting key shares..."
count=0
jq -r '.shares[]' "$SHARES_PATH" | head -5 | while read -r share; do
  ((count++))
  echo "→ Submitting share $count..."
  buf curl \
    --schema "$API_PATH" \
    --protocol grpc \
    --cacert "$CA_CERT" \
    --data "{\"key_piece\": \"$share\"}" \
    --header "authorization: Bearer $GK_TOKEN" \
    --header "authority: $SERVER_HOST" \
    "https://$SERVER_HOST:$SERVER_PORT/gophkeeper.v1.AdminService/Unseal"
done

echo "🎉 ✅ Unseal complete."