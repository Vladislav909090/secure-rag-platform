package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
)

// Chunk представляет сегмент документа с эмбеддингом.
type Chunk struct {
	DocumentUUID       string
	ChunkIndex         int32
	ChunkText          string
	Embedding          pgvector.Vector
	EmbeddingModel     string
	EmbeddingDimension int32
}

// ChunkMatch возвращается при поиске по вектору.
type ChunkMatch struct {
	DocumentUUID string
	ChunkIndex   int32
	ChunkText    string
	Distance     float32
}

// DeleteChunks удаляет все сегменты документа.
func (r *Repo) DeleteChunks(ctx context.Context, documentUUID string) error {
	query := `DELETE FROM rag_chunks WHERE document_uuid = $1`
	_, err := r.pool.Exec(ctx, query, documentUUID)
	return err
}

// InsertChunks добавляет набор сегментов.
func (r *Repo) InsertChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, chunk := range chunks {
		if strings.TrimSpace(chunk.EmbeddingModel) == "" {
			return fmt.Errorf("embedding model is required")
		}

		var chunkID int64
		query := `
			INSERT INTO rag_chunks (document_uuid, chunk_index, chunk_text)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (document_uuid, chunk_index)
			 DO UPDATE SET chunk_text = EXCLUDED.chunk_text
			 RETURNING id
		`
		insertChunkErr := tx.QueryRow(
			ctx,
			query,
			chunk.DocumentUUID,
			chunk.ChunkIndex,
			chunk.ChunkText,
		).Scan(&chunkID)
		if insertChunkErr != nil {
			return fmt.Errorf("insert chunk: %w", insertChunkErr)
		}

		query = `
			INSERT INTO rag_embeddings (chunk_id, embedding, embedding_model, embedding_dimension)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (chunk_id, embedding_model)
			 DO UPDATE SET
			   embedding = EXCLUDED.embedding,
			   embedding_dimension = EXCLUDED.embedding_dimension
		`
		_, insertEmbeddingErr := tx.Exec(
			ctx,
			query,
			chunkID,
			chunk.Embedding,
			chunk.EmbeddingModel,
			chunk.EmbeddingDimension,
		)
		if insertEmbeddingErr != nil {
			return fmt.Errorf("insert embedding: %w", insertEmbeddingErr)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// SearchSimilar ищет ближайшие сегменты по вектору.
func (r *Repo) SearchSimilar(
	ctx context.Context,
	vector pgvector.Vector,
	embeddingModel string,
	embeddingDimension int32,
	indexedDimension int32,
	topK int32,
	documentUUIDs []string,
) ([]ChunkMatch, error) {
	if topK <= 0 {
		return nil, nil
	}
	if len(documentUUIDs) == 0 {
		return nil, nil
	}

	embeddingModel = strings.TrimSpace(embeddingModel)
	if embeddingModel == "" {
		return nil, fmt.Errorf("embedding model is required")
	}
	if embeddingDimension <= 0 {
		return nil, fmt.Errorf("embedding dimension is required")
	}

	distanceExpr := "e.embedding <=> $1"
	dimensionFilter := "e.embedding_dimension = $3"
	topKPlaceholder := "$4"
	documentUUIDsPlaceholder := "$5"
	args := []any{vector, embeddingModel, embeddingDimension, topK}

	if indexedDimension > 0 && embeddingDimension == indexedDimension {
		dim := int(indexedDimension)
		distanceExpr = fmt.Sprintf("(e.embedding::vector(%d) <=> $1::vector(%d))", dim, dim)
		dimensionFilter = fmt.Sprintf("e.embedding_dimension = %d", dim)
		topKPlaceholder = "$3"
		documentUUIDsPlaceholder = "$4"
		args = []any{vector, embeddingModel, topK}
	}

	where := ""
	if len(documentUUIDs) > 0 {
		where = " AND c.document_uuid = ANY(" + documentUUIDsPlaceholder + ")"
		args = append(args, documentUUIDs)
	}

	query := strings.Join([]string{
		"SELECT c.document_uuid, c.chunk_index, c.chunk_text,",
		"  " + distanceExpr + " AS distance",
		"FROM rag_chunks c",
		"JOIN rag_embeddings e ON e.chunk_id = c.id",
		"WHERE e.embedding_model = $2",
		"  AND " + dimensionFilter,
		where,
		"ORDER BY " + distanceExpr,
		"LIMIT " + topKPlaceholder,
	}, " ")

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	matches := make([]ChunkMatch, 0)
	for rows.Next() {
		var match ChunkMatch
		var distance float64
		if scanErr := rows.Scan(
			&match.DocumentUUID,
			&match.ChunkIndex,
			&match.ChunkText,
			&distance,
		); scanErr != nil {
			return nil, fmt.Errorf("scan chunk: %w", scanErr)
		}
		match.Distance = float32(distance)
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chunks: %w", err)
	}

	return matches, nil
}
