package repository

import (
	"context"
	"fmt"
	"time"
)

// UpdateDocument обновляет title и/или description.
func (r *Repo) UpdateDocument(
	ctx context.Context,
	uuid string,
	title *string,
	description *string,
	updatedAt time.Time,
) error {
	query := `
		UPDATE documents
		SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			updated_at = $3
		WHERE uuid = $4
		  AND deleted_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, title, description, updatedAt, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
