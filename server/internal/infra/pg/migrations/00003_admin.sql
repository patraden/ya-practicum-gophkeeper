-- +goose Up
-- +goose StatementBegin
CREATE TABLE rek (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    rek_hash BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rek;
-- +goose StatementEnd
