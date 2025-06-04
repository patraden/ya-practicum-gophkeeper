-- name: CreateUser :one
INSERT INTO users (id, username, role, password, salt, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    role = users.role,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, role, password, salt, created_at, updated_at;

-- name: GetUser :one
SELECT id, username, role, password, salt, created_at, updated_at
FROM users
WHERE username = $1;