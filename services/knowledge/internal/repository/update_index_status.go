package repository

import "context"

// UpdateIndexStatus обновляет статус индексации версии.
func (r *Repo) UpdateIndexStatus(ctx context.Context, versionID int64, status string) error {
	query := `
		UPDATE document_versions
		SET index_status = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, status, versionID)

	return err
}
