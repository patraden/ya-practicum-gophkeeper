-- +goose Up
-- +goose StatementBegin
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE test_table;
-- +goose StatementEnd
