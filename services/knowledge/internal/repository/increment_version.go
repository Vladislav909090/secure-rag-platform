package repository

import (
	"context"
	"time"
)

// IncrementVersion обновляет current_version_number документа.
func (r *Repo) IncrementVersion(ctx context.Context, docID int64, newVersion int32, updatedAt time.Time) error {
	query := `
		UPDATE documents
		SET
			current_version_number = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.pool.Exec(ctx, query, newVersion, updatedAt, docID)

	return err
}
