-- +goose Up
DROP INDEX IF EXISTS rag_embeddings_embedding_hnsw_idx;

ALTER TABLE rag_embeddings
    ALTER COLUMN embedding TYPE VECTOR;

CREATE INDEX IF NOT EXISTS rag_embeddings_model_dimension_idx
    ON rag_embeddings (embedding_model, embedding_dimension);

-- +goose Down
DROP INDEX IF EXISTS rag_embeddings_embedding_hnsw_768_idx;
DROP INDEX IF EXISTS rag_embeddings_model_dimension_idx;

ALTER TABLE rag_embeddings
    ALTER COLUMN embedding TYPE VECTOR(768);

CREATE INDEX IF NOT EXISTS rag_embeddings_embedding_hnsw_idx
    ON rag_embeddings
    USING hnsw (embedding vector_cosine_ops);
