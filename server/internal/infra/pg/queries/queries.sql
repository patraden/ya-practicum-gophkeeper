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

-- name: GetUserByID :one
SELECT id, username, role, created_at, updated_at, password, salt, verifier, bucket_name, identity_id
FROM users
WHERE id = $1;

-- name: CreateIdentityToken :exec
INSERT INTO user_identity_tokens (
    user_id,
    access_token,
    refresh_token,
    expires_at,
    refresh_expires_at,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id) DO UPDATE
SET 
  access_token = $2,
  refresh_token = $3,
  expires_at = $4,
  refresh_expires_at = $5,
  created_at = $6,
  updated_at = $7;

-- name: GetIdentityToken :one
SELECT 
    user_id,
    access_token,
    refresh_token,
    expires_at,
    refresh_expires_at,
    created_at,
    updated_at
FROM user_identity_tokens
WHERE user_id = $1;

-- name: DeleteIdentityToken :exec
DELETE FROM user_identity_tokens
WHERE user_id = $1;

-- name: CreateREKHash :exec
INSERT INTO rek (rek_hash)
VALUES ($1);

-- name: GetREKHash :one
SELECT rek_hash, created_at
FROM rek
LIMIT 1;

-- name: CreateSecret :exec
INSERT INTO secrets (user_id, secret_id, secret_name, current_version_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateSecret :exec
UPDATE secrets
SET current_version_id = $3,
    updated_at = NOW()
WHERE user_id = $1 AND secret_id = $2;

-- name: CreateSecretInitRequest :one
WITH candidate(parent_version_id) AS (
  -- Case: existing secret with matching parent
  SELECT current_version_id AS parent_version_id
  FROM secrets
  WHERE user_id = $1 AND secret_id = $2 AND COALESCE(secrets.current_version_id, $6) = $6
  UNION ALL
  -- Case: new secret
  SELECT NULL::UUID AS parent_version_id
  WHERE NOT EXISTS (
    SELECT 1 FROM secrets WHERE user_id = $1 AND secret_id = $2
  )
)
INSERT INTO secret_requests_in_progress (
  user_id,
  secret_id,
  secret_name,
  s3_url,
  version_id,
  parent_version_id,
  request_type,
  token,
  client_info,
  secret_size,
  secret_hash,
  secret_dek,
  meta,
  created_at,
  expires_at
)
SELECT
  $1, $2, $3, $4, $5, candidate.parent_version_id, $7, $8,
  $9, $10, $11, $12, $13, $14, $15
FROM candidate
ON CONFLICT (user_id, secret_id) DO UPDATE
  SET user_id = EXCLUDED.user_id
RETURNING 
  user_id,
  secret_id,
  secret_name,
  s3_url,
  version_id,
  parent_version_id,
  request_type,
  token,
  client_info,
  secret_size,
  secret_hash,
  secret_dek,
  meta,
  created_at,
  expires_at;

-- name: DeleteSecretInitRequest :exec
DELETE FROM secret_requests_in_progress
WHERE user_id = $1 AND secret_id = $2;

-- name: CreateSecretCommitRequest :exec
INSERT INTO secret_requests_completed (
    user_id,
    secret_id,
    s3_url,
    version_id,
    parent_version_id,
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
    $9, $10, $11, $12, $13, $14, $15, $16
);