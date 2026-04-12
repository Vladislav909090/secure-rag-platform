package repository

import (
	"context"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateVersion вставляет версию документа и возвращает сгенерированный id.
func (r *Repo) CreateVersion(ctx context.Context, v *model.DocumentVersion) error {
	query := `
		INSERT INTO document_versions (
			uuid,
			document_id,
			version_number,
			file_name,
			file_extension,
			mime_type,
			size_bytes,
			checksum_sha256,
			storage_key,
			index_status,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		v.UUID, v.DocumentID, v.VersionNumber, v.FileName, v.FileExtension,
		v.MimeType, v.SizeBytes, v.ChecksumSHA256, v.StorageKey, v.IndexStatus, v.CreatedAt,
	).Scan(&v.ID)

	return err
}
