package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateDocument вставляет документ и возвращает сгенерированный id
func (r *Repo) CreateDocument(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (
			uuid,
			title,
			description,
			attributes,
			file_name,
			file_extension,
			mime_type,
			size_bytes,
			checksum_sha256,
			storage_key,
			index_status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	attrJSON, err := json.Marshal(doc.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		doc.UUID, doc.Title, doc.Description, attrJSON,
		doc.FileName, doc.FileExtension, doc.MimeType,
		doc.SizeBytes, doc.ChecksumSHA256, doc.StorageKey, doc.IndexStatus,
		doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)

	return err
}
