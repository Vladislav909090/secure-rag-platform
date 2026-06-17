package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidRoleCode = errors.New("invalid role code")
)

// Repo реализует операции персистентности IAM
type Repo struct {
	pool *pgxpool.Pool
}

// NewRepo создает новый экземпляр Repo
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func normalizeAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return map[string]any{}
	}

	return attrs
}

func toJSON(attrs map[string]any) ([]byte, error) {
	normalized := normalizeAttributes(attrs)
	b, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal attributes: %w", err)
	}

	return b, nil
}

func fromJSON(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal attributes: %w", err)
	}

	return out, nil
}
