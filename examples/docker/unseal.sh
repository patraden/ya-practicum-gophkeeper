export GK_TOKEN="Bearer $(cat .token)"

buf curl \
  --schema ./api \
  --protocol grpc \
  --cacert .certs/ca.cert \
  --data '{"key_piece": "" }' \
  --header "authorization: ${GK_TOKEN}" \
  --header "authority: localhost" \
  https://localhost:3200/gophkeeper.v1.AdminService/Unseal