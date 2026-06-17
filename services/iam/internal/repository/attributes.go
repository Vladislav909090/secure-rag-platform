package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetUserAttributes возвращает объект атрибутов пользователя
func (r *Repo) GetUserAttributes(ctx context.Context, userID string) (map[string]any, error) {
	query := `
		SELECT attributes
		FROM user_attributes
		WHERE user_id = $1
	`

	var raw []byte
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return map[string]any{}, nil
		}

		return nil, fmt.Errorf("query user attributes: %w", err)
	}

	attrs, err := fromJSON(raw)
	if err != nil {
		return nil, err
	}

	return attrs, nil
}

// ReplaceUserAttributes полностью заменяет атрибуты пользователя
func (r *Repo) ReplaceUserAttributes(
	ctx context.Context,
	userID string,
	attrs map[string]any,
	updatedBy *string,
) (map[string]any, error) {
	attrsJSON, err := toJSON(attrs)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO user_attributes (
			user_id,
			attributes,
			updated_at,
			updated_by
		)
		VALUES ($1, $2::jsonb, NOW(), $3)
		ON CONFLICT (user_id) DO UPDATE
		SET
			attributes = EXCLUDED.attributes,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
		RETURNING attributes
	`

	var raw []byte
	err = r.pool.QueryRow(ctx, query, userID, attrsJSON, updatedBy).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("replace user attributes: %w", err)
	}

	updated, err := fromJSON(raw)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// DeleteUserAttributeKey удаляет один ключ из атрибутов пользователя
func (r *Repo) DeleteUserAttributeKey(
	ctx context.Context,
	userID string,
	key string,
	updatedBy *string,
) (map[string]any, error) {
	query := `
		INSERT INTO user_attributes (
			user_id,
			attributes,
			updated_at,
			updated_by
		)
		VALUES ($1, '{}'::jsonb, NOW(), $3)
		ON CONFLICT (user_id) DO UPDATE
		SET
			attributes = user_attributes.attributes - $2,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
		RETURNING attributes
	`

	var raw []byte
	if err := r.pool.QueryRow(ctx, query, userID, key, updatedBy).Scan(&raw); err != nil {
		return nil, fmt.Errorf("delete user attribute key: %w", err)
	}

	updated, err := fromJSON(raw)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
