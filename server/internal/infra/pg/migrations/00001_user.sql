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
  verifier BYTEA NOT NULL
);

CREATE TABLE keys (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  kek BYTEA NOT NULL,
  algorithm VARCHAR(10) NOT NULL DEFAULT 'aes-gcm',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE keys;
-- +goose StatementEnd
