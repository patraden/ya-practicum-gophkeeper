-- name: CreateUser :one
INSERT INTO "user" (id, username, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (username) DO UPDATE
SET id = "user".id,
    username = "user".username,
    password = "user".password,
    role = "user".role,
    created_at = "user".created_at,
    updated_at = "user".updated_at
RETURNING id, username, password, role, created_at, updated_at;