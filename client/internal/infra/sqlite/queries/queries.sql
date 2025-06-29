-- name: CreateUser :exec
INSERT INTO users (id, username, verifier, role, salt, bucketname, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateUserToken :exec
INSERT INTO users_server_tokens (user_id, token, ttl)
VALUES (?, ?, ?);