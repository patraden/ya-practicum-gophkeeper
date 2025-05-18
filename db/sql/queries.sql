-- name: CreateUser :one
INSERT INTO users (id, username, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    username = users.username,
    password = users.password,
    role = users.role,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, password, role, created_at, updated_at;