-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
  id TEXT PRIMARY KEY CHECK (length(id) = 36),
  username TEXT UNIQUE NOT NULL CHECK (length(username) <= 64),
  verifier BLOB NOT NULL,
  role TEXT NOT NULL,
  salt BLOB NOT NULL,
  bucketname TEXT NOT NULL CHECK (length(bucketname) <= 32),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL
);

CREATE TABLE users_server_tokens (
  user_id TEXT PRIMARY KEY CHECK (length(user_id) = 36) REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL,
  ttl INTEGER NOT NULL  -- token TTL in seconds
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS users_server_tokens;
-- +goose StatementEnd
