package repository

import (
	"context"
	"errors"
	"fmt"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/jackc/pgx/v5"
)

// CreateUserInput содержит данные, необходимые для создания пользователя
type CreateUserInput struct {
	Login        string
	PasswordHash string
	IsActive     bool
	RoleCodes    []string
	Attributes   map[string]any
	CreatedBy    *string
}

// UpdateUserInput содержит данные частичного обновления пользователя
type UpdateUserInput struct {
	UserID       string
	Login        *string
	PasswordHash *string
	IsActive     *bool
}

func scanUser(row pgx.Row) (*model.User, error) {
	var user model.User
	if err := row.Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.IsActive,
		&user.CtxVer,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID возвращает пользователя по UUID
func (r *Repo) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	query := `
		SELECT
			id,
			login,
			password_hash,
			is_active,
			ctx_ver,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return user, nil
}

// GetUserByLogin возвращает пользователя по уникальному логину
func (r *Repo) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
		SELECT
			id,
			login,
			password_hash,
			is_active,
			ctx_ver,
			created_at,
			updated_at
		FROM users
		WHERE login = $1
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, login))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("query user by login: %w", err)
	}

	return user, nil
}

// ListUsers возвращает всех пользователей, отсортированных по времени создания
func (r *Repo) ListUsers(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT
			id,
			login,
			password_hash,
			is_active,
			ctx_ver,
			created_at,
			updated_at
		FROM users
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]*model.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

// CreateUser создает пользователя вместе с ролями и атрибутами в одной транзакции
func (r *Repo) CreateUser(ctx context.Context, input CreateUserInput) (*model.User, error) {
	var created *model.User

	if err := withTx(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			INSERT INTO users (
				login,
				password_hash,
				is_active
			)
			VALUES ($1, $2, $3)
			RETURNING
				id,
				login,
				password_hash,
				is_active,
				ctx_ver,
				created_at,
				updated_at
		`

		user, err := scanUser(tx.QueryRow(
			ctx,
			query,
			input.Login,
			input.PasswordHash,
			input.IsActive,
		))
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}

		attrsJSON, err := toJSON(input.Attributes)
		if err != nil {
			return err
		}

		query = `
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
		`

		if _, err := tx.Exec(ctx, query, user.ID, attrsJSON, input.CreatedBy); err != nil {
			return fmt.Errorf("upsert user attributes: %w", err)
		}

		roleCodes := normalizeRoleCodes(input.RoleCodes)
		if err := setUserRolesTx(ctx, tx, user.ID, roleCodes, input.CreatedBy); err != nil {
			return err
		}

		created = user

		return nil
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateUser обновляет изменяемые поля пользователя
func (r *Repo) UpdateUser(ctx context.Context, input UpdateUserInput) (*model.User, error) {
	var login any
	if input.Login != nil {
		login = *input.Login
	}

	var passwordHash any
	if input.PasswordHash != nil {
		passwordHash = *input.PasswordHash
	}

	var isActive any
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `
		UPDATE users
		SET
			login = COALESCE($2::text, login),
			password_hash = COALESCE($3::text, password_hash),
			is_active = COALESCE($4::boolean, is_active),
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			login,
			password_hash,
			is_active,
			ctx_ver,
			created_at,
			updated_at
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, input.UserID, login, passwordHash, isActive))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// IncrementContextVersion увеличивает ctx_ver и возвращает новое значение
func (r *Repo) IncrementContextVersion(ctx context.Context, userID string) (int64, error) {
	query := `
		UPDATE users
		SET
			ctx_ver = ctx_ver + 1,
			updated_at = NOW()
		WHERE id = $1
		RETURNING ctx_ver
	`

	var ctxVer int64
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&ctxVer); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, fmt.Errorf("increment ctx_ver: %w", err)
	}

	return ctxVer, nil
}
