package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

// ListActiveDocuments возвращает все активные (не удалённые) документы.
func (r *Repo) ListActiveDocuments(ctx context.Context) ([]*model.Document, error) {
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
		WHERE deleted_at IS NULL
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc := &model.Document{}
		var attrJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.UUID, &doc.Title, &doc.Description, &attrJSON,
			&doc.CurrentVersionNumber, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
		); err != nil {
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
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}
