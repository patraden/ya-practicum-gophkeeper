-- name: CreateUser :one
INSERT INTO users (id, username, role, created_at, updated_at, password, salt, verifier)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    role = users.role,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, role, created_at, updated_at, password, salt, verifier;

-- name: CreateUserKey :exec
INSERT INTO keys (user_id, kek, algorithm, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUser :one
SELECT id, username, role, created_at, updated_at, password, salt, verifier
FROM users
WHERE username = $1;

-- name: CreateREKHash :exec
INSERT INTO rek (rek_hash)
VALUES ($1);

-- name: GetREKHash :one
SELECT rek_hash, created_at
FROM rek
LIMIT 1;