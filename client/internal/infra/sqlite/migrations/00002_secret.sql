-- +goose Up
-- +goose StatementBegin
CREATE TABLE secrets (
    user_id             TEXT NOT NULL CHECK (length(user_id) = 36),
    secret_id           TEXT NOT NULL CHECK (length(secret_id) = 36),
    secret_name         TEXT NOT NULL CHECK (length(secret_name) <= 64),
    version_id          TEXT NOT NULL CHECK (length(version_id) = 36),
    parent_version_id   TEXT CHECK (parent_version_id IS NULL OR length(parent_version_id) = 36),
    file_path           TEXT NOT NULL,
    secret_size         INTEGER NOT NULL,
    secret_hash         BLOB NOT NULL,
    secret_dek          BLOB NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL,
    in_sync             INTEGER NOT NULL DEFAULT 0,  -- 0 = false, 1 = true

    PRIMARY KEY (user_id, secret_id),
    UNIQUE (user_id, secret_name)
);

CREATE TABLE secret_meta (
    user_id      TEXT NOT NULL CHECK (length(user_id) = 36),
    secret_id    TEXT NOT NULL CHECK (length(secret_id) = 36),
    meta         TEXT NOT NULL DEFAULT '{}',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL,
    FOREIGN KEY (user_id, secret_id)
        REFERENCES secrets(user_id, secret_id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secret_meta;
DROP TABLE IF EXISTS secrets;
-- +goose StatementEnd
