-- name: CreateUser :one
INSERT INTO users (id, username, verifier, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, username, verifier, role, created_at, updated_at;

-- name: CountUser :one
SELECT count(*) FROM users;