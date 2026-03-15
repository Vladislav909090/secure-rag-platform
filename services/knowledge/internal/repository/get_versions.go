package repository

import (
	"context"

	"secure-rag-platform/services/knowledge/internal/model"
)

// GetVersionsByDocumentID возвращает все версии документа.
func (r *Repo) GetVersionsByDocumentID(ctx context.Context, docID int64) ([]*model.DocumentVersion, error) {
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
		ORDER BY version_number
	`

	rows, err := r.pool.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*model.DocumentVersion
	for rows.Next() {
		v := &model.DocumentVersion{}

		if err := rows.Scan(
			&v.ID, &v.UUID, &v.DocumentID, &v.VersionNumber,
			&v.FileName, &v.FileExtension, &v.MimeType, &v.SizeBytes,
			&v.ChecksumSHA256, &v.StorageKey, &v.IndexStatus, &v.CreatedAt,
		); err != nil {
			return nil, err
		}

		versions = append(versions, v)
	}

	return versions, rows.Err()
}
