package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxIndexedEmbeddingDimension = 2000

// Repo обеспечивает доступ к векторному хранилищу
type Repo struct {
	pool *pgxpool.Pool
}

// NewRepo создаёт новый репозиторий
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// EnsureEmbeddingIndex создает HNSW-индекс для выбранной размерности embeddings
func (r *Repo) EnsureEmbeddingIndex(ctx context.Context, dimension int32) error {
	if dimension <= 0 {
		return fmt.Errorf("indexed embedding dimension is required")
	}
	if dimension > maxIndexedEmbeddingDimension {
		return fmt.Errorf("indexed embedding dimension exceeds pgvector vector limit: %d", dimension)
	}

	dim := int(dimension)
	query := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS rag_embeddings_embedding_hnsw_%d_idx
		ON rag_embeddings
		USING hnsw ((embedding::vector(%d)) vector_cosine_ops)
		WHERE embedding_dimension = %d
	`, dim, dim, dim)

	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("create embedding hnsw index: %w", err)
	}

	return nil
}
