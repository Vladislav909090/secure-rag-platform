-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS rag_chunks (
    id BIGSERIAL PRIMARY KEY,
    document_uuid TEXT NOT NULL,
    version_number INT NOT NULL,
    chunk_index INT NOT NULL,
    chunk_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_uuid, version_number, chunk_index)
);

CREATE INDEX IF NOT EXISTS rag_chunks_document_idx
    ON rag_chunks (document_uuid, version_number);

CREATE TABLE IF NOT EXISTS rag_embeddings (
    id BIGSERIAL PRIMARY KEY,
    chunk_id BIGINT NOT NULL REFERENCES rag_chunks(id) ON DELETE CASCADE,
    embedding VECTOR(768) NOT NULL,
    embedding_model TEXT NOT NULL,
    embedding_dimension INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chunk_id, embedding_model)
);

CREATE INDEX IF NOT EXISTS rag_embeddings_model_idx
    ON rag_embeddings (embedding_model);

CREATE INDEX IF NOT EXISTS rag_embeddings_embedding_hnsw_idx
    ON rag_embeddings
    USING hnsw (embedding vector_cosine_ops);

-- +goose Down
DROP INDEX IF EXISTS rag_embeddings_embedding_hnsw_idx;
DROP INDEX IF EXISTS rag_embeddings_model_idx;
DROP INDEX IF EXISTS rag_chunks_document_idx;
DROP TABLE IF EXISTS rag_embeddings;
DROP TABLE IF EXISTS rag_chunks;
