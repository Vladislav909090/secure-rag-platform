package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// UpdateAttributes полностью заменяет attributes документа.
func (r *Repo) UpdateAttributes(ctx context.Context, uuid string, attributes map[string]any, updatedAt time.Time) error {
	query := `
		UPDATE documents
		SET
			attributes = $1,
			updated_at = $2
		WHERE uuid = $3
		  AND deleted_at IS NULL
	`

	attrJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query, attrJSON, updatedAt, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
