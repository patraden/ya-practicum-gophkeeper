-- +goose Up
-- +goose StatementBegin
CREATE TYPE request_type AS ENUM ('put', 'get');
CREATE TYPE request_status AS ENUM ('completed', 'aborted', 'expired', 'cancelled');
CREATE TYPE request_commiter AS ENUM ('user', 'server', 's3');

CREATE TABLE secrets (
    user_id         UUID NOT NULL,
    secret_id       UUID NOT NULL,
    secret_name     VARCHAR(64) NOT NULL,
    current_version UUID NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, secret_id),
    UNIQUE (user_id, secret_name)
);

CREATE TABLE secret_versions (
    id             BIGSERIAL PRIMARY KEY,
    user_id        UUID NOT NULL,
    secret_id      UUID NOT NULL,
    version        UUID NOT NULL,
    parent_version UUID,
    s3_url         TEXT NOT NULL,
    secret_size    BIGINT NOT NULL,
    secret_hash    BYTEA NOT NULL,
    secret_dek     BYTEA NOT NULL,
    created_at     TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id, secret_id) REFERENCES secrets(user_id, secret_id) ON DELETE CASCADE
);

CREATE INDEX idx_secret_versions_user_secret ON secret_versions(user_id, secret_id);

CREATE TABLE secret_meta (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL,
    secret_id   UUID NOT NULL,
    meta        JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id, secret_id) REFERENCES secrets(user_id, secret_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uniq_secrets_meta ON secret_meta(user_id, secret_id);

CREATE TABLE secret_requests_in_progress (
    id                BIGSERIAL PRIMARY KEY,
    user_id           UUID NOT NULL,
    secret_id         UUID NOT NULL,
    secret_name       VARCHAR(64) NOT NULL,
    s3_url            TEXT NOT NULL,
    version           UUID NOT NULL,
    parent_version    UUID NULL,
    request_type      request_type NOT NULL,
    token             BIGINT NOT NULL,
    client_info       VARCHAR(128) NOT NULL,
    secret_size       BIGINT NOT NULL,
    secret_hash       BYTEA,
    secret_dek        BYTEA,
    meta              JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at        TIMESTAMP NOT NULL,
    UNIQUE (user_id, secret_id)
);

CREATE TABLE secret_requests_completed (
    id                BIGSERIAL PRIMARY KEY,
    user_id           UUID NOT NULL,
    secret_id         UUID NOT NULL,
    s3_url            TEXT NOT NULL,
    version           UUID NOT NULL,
    parent_version    UUID NULL,
    request_type      request_type NOT NULL,
    token             BIGINT NOT NULL,
    client_info       VARCHAR(128) NOT NULL,
    secret_size       BIGINT NOT NULL,
    secret_hash       BYTEA,
    secret_dek        BYTEA,
    created_at        TIMESTAMP NOT NULL,
    expires_at        TIMESTAMP NOT NULL,
    finished_at       TIMESTAMP NOT NULL,
    status            request_status NOT NULL,
    commited_by       request_commiter NOT NULL
);

CREATE UNIQUE INDEX uniq_secret_requests_completed ON secret_requests_completed(user_id, secret_id, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS secret_requests_completed;
DROP TABLE IF EXISTS secret_requests_issued;
DROP INDEX IF EXISTS uniq_secrets_meta;
DROP TABLE IF EXISTS secret_meta;
DROP TABLE IF EXISTS secret_versions;
DROP TABLE IF EXISTS secrets;
DROP TYPE IF EXISTS request_commiter;
DROP TYPE IF EXISTS request_status;
DROP TYPE IF EXISTS request_type;

-- +goose StatementEnd