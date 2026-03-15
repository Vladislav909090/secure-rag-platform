package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"secure-rag-platform/services/knowledge/internal/model"
)

// GetDocumentByUUID возвращает документ по uuid. Возвращает nil, nil если не найден.
func (r *Repo) GetDocumentByUUID(ctx context.Context, uuid string) (*model.Document, error) {
	query := `
		SELECT
			id,
			uuid,
			title,
			description,
			attributes,
			current_version_number,
			created_at,
			updated_at,
			deleted_at
		FROM documents
		WHERE uuid = $1
	`

	doc := &model.Document{}
	var attrJSON []byte

	err := r.pool.QueryRow(ctx, query, uuid).Scan(
		&doc.ID, &doc.UUID, &doc.Title, &doc.Description, &attrJSON,
		&doc.CurrentVersionNumber, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	if len(attrJSON) > 0 {
		if err := json.Unmarshal(attrJSON, &doc.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}
	if doc.Attributes == nil {
		doc.Attributes = map[string]any{}
	}

	return doc, nil
}
