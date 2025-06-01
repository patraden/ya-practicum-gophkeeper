-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
  id UUID PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  role VARCHAR(5) NOT NULL,
  password BYTEA NOT NULL,
  salt BYTEA NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "user" 
-- +goose StatementEnd
