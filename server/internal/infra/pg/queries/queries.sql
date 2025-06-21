-- name: CreateUser :one
INSERT INTO users (id, username, role, created_at, updated_at, password, salt, verifier, bucket_name, identity_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    role = users.role,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, role, created_at, updated_at, password, salt, verifier, bucket_name, identity_id;

-- name: CreateUserKey :exec
INSERT INTO user_crypto_keys (user_id, kek, algorithm, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUser :one
SELECT id, username, role, created_at, updated_at, password, salt, verifier, bucket_name, identity_id
FROM users
WHERE username = $1;

-- name: CreateIdentityToken :exec
INSERT INTO user_identity_tokens (
    user_id,
    access_token,
    refresh_token,
    expires_at,
    refresh_expires_at,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CreateREKHash :exec
INSERT INTO rek (rek_hash)
VALUES ($1);

-- name: GetREKHash :one
SELECT rek_hash, created_at
FROM rek
LIMIT 1;

-- name: CreateSecret :exec
INSERT INTO secrets (user_id, secret_id, secret_name, current_version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateSecret :exec
UPDATE secrets
SET current_version = $3,
    updated_at = NOW()
WHERE user_id = $1 AND secret_id = $2;

-- name: InsertSecretRequestIssued :one
INSERT INTO secret_requests_issued (
    user_id,
    secret_id,
    version,
    parent_version,
    request_type,
    token,
    client_info,
    secret_size,
    secret_hash,
    secret_dek,
    expires_at
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11
)
ON CONFLICT (user_id, secret_id) DO NOTHING
RETURNING *;

-- name: DeleteSecretRequestIssued :exec
DELETE FROM secret_requests_issued
WHERE user_id = $1 AND secret_id = $2;

-- name: InsertSecretRequestCompleted :exec
INSERT INTO secret_requests_completed (
    user_id,
    secret_id,
    version,
    parent_version,
    request_type,
    token,
    client_info,
    secret_size,
    secret_hash,
    secret_dek,
    created_at,
    expires_at,
    finished_at,
    status,
    commited_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15
);