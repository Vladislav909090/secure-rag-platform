package repository

import "context"

// UpdateIndexStatus обновляет статус индексации документа.
func (r *Repo) UpdateIndexStatus(ctx context.Context, docID int64, status string) error {
	query := `
		UPDATE documents
		SET
			index_status = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, status, docID)

	return err
}
