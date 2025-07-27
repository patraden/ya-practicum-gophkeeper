-- name: CreateUser :exec
INSERT INTO users (id, username, verifier, role, salt, bucketname, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetUser :one
SELECT
    id,
    username,
    verifier,
    role,
    salt,
    bucketname,
    created_at,
    updated_at
FROM users
WHERE username = ?;

-- name: CreateUserToken :exec
INSERT INTO users_server_tokens (user_id, token, ttl)
VALUES (?, ?, ?);

-- name: CreateSecret :exec
INSERT INTO secrets (
    user_id,
    secret_id,
    secret_name,
    version_id,
    parent_version_id,
    file_path,
    secret_size,
    secret_hash,
    secret_dek,
    created_at,
    updated_at,
    in_sync
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateSecret :exec
UPDATE secrets
SET
    version_id = ?,
    parent_version_id = ?,
    file_path = ?,
    secret_size = ?,
    secret_hash = ?,
    secret_dek = ?,
    updated_at = ?,
    in_sync = ?
WHERE user_id = ? AND secret_id = ?;

-- name: GetSecret :one
SELECT
    secrets.user_id,
    secrets.secret_id,
    secrets.secret_name,
    secrets.version_id,
    secrets.parent_version_id,
    secrets.file_path,
    secrets.secret_size,
    secrets.secret_hash,
    secrets.secret_dek,
    secrets.created_at,
    secrets.updated_at,
    secrets.in_sync
FROM secrets
JOIN users ON users.id = secrets.user_id
WHERE users.username = ? AND secret_name = ?;