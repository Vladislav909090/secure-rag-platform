package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateDocument вставляет документ и возвращает сгенерированный id.
func (r *Repo) CreateDocument(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (
			uuid,
			title,
			description,
			attributes,
			current_version_number,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	attrJSON, err := json.Marshal(doc.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		doc.UUID, doc.Title, doc.Description, attrJSON,
		doc.CurrentVersionNumber, doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)

	return err
}
