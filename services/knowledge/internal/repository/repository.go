package repository

import "github.com/jackc/pgx/v5/pgxpool"

// Repo — репозиторий для работы с документами.
type Repo struct {
	pool *pgxpool.Pool
}

// NewRepo создаёт новый репозиторий.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}
