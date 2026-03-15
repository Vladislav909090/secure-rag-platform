-- +goose Up
CREATE TABLE documents (
    id                     BIGSERIAL    PRIMARY KEY,
    uuid                   UUID         NOT NULL UNIQUE,
    title                  TEXT         NOT NULL,
    description            TEXT,
    attributes             JSONB        NOT NULL DEFAULT '{}'::jsonb,
    current_version_number INT          NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ  NOT NULL,
    updated_at             TIMESTAMPTZ  NOT NULL,
    deleted_at             TIMESTAMPTZ
);

CREATE TABLE document_versions (
    id              BIGSERIAL   PRIMARY KEY,
    uuid            UUID        NOT NULL UNIQUE,
    document_id     BIGINT      NOT NULL,
    version_number  INT         NOT NULL,
    file_name       TEXT        NOT NULL,
    file_extension  TEXT        NOT NULL,
    mime_type       TEXT        NOT NULL,
    size_bytes      BIGINT      NOT NULL,
    checksum_sha256 TEXT        NOT NULL,
    storage_key     TEXT        NOT NULL,
    index_status    TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    UNIQUE (document_id, version_number)
);

-- +goose Down
DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS documents;