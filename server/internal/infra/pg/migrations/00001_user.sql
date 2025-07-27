-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  role integer NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP,
  password BYTEA NOT NULL,
  salt BYTEA NOT NULL,
  verifier BYTEA NOT NULL,
  bucket_name VARCHAR(32) NOT NULL,
  identity_id VARCHAR(36) NOT NULL
);

CREATE TABLE user_crypto_keys (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  kek BYTEA NOT NULL,
  algorithm VARCHAR(10) NOT NULL DEFAULT 'aes-gcm',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE user_identity_tokens (
    user_id UUID PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    refresh_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_identity_tokens;
DROP TABLE user_crypto_keys;
DROP TABLE users;
-- +goose StatementEnd
