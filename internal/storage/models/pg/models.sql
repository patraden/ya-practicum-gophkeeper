-- name: CreateUser :one
INSERT INTO "user" (id, username, role, password, salt, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (username) DO UPDATE
SET id = "user".id,
    role = "user".role,
    created_at = "user".created_at,
    updated_at = "user".updated_at
RETURNING id, username, role, password, salt, created_at, updated_at;