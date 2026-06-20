package repository

import (
	"context"
	"fmt"
	"time"
)

// SoftDeleteDocument мягко удаляет документ
func (r *Repo) SoftDeleteDocument(ctx context.Context, uuid string, deletedAt time.Time) error {
	query := `
		UPDATE documents
		SET
			deleted_at = $1,
			updated_at = $1
		WHERE uuid = $2
		  AND deleted_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, deletedAt, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
