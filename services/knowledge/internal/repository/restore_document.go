package repository

import (
	"context"
	"fmt"
	"time"
)

// RestoreDocument убирает soft delete.
func (r *Repo) RestoreDocument(ctx context.Context, uuid string, updatedAt time.Time) error {
	query := `
		UPDATE documents
		SET
			deleted_at = NULL,
			updated_at = $1
		WHERE uuid = $2
		  AND deleted_at IS NOT NULL
	`

	tag, err := r.pool.Exec(ctx, query, updatedAt, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
