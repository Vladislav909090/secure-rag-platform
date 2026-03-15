package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"secure-rag-platform/services/knowledge/internal/model"
)

// GetVersion возвращает конкретную версию. Возвращает nil, nil если не найдена.
func (r *Repo) GetVersion(ctx context.Context, docID int64, versionNumber int32) (*model.DocumentVersion, error) {
	query := `
		SELECT
			id,
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
		FROM document_versions
		WHERE document_id = $1
		  AND version_number = $2
	`

	v := &model.DocumentVersion{}

	err := r.pool.QueryRow(ctx, query, docID, versionNumber).Scan(
		&v.ID, &v.UUID, &v.DocumentID, &v.VersionNumber,
		&v.FileName, &v.FileExtension, &v.MimeType, &v.SizeBytes,
		&v.ChecksumSHA256, &v.StorageKey, &v.IndexStatus, &v.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return v, nil
}
