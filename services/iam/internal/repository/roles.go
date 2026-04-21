package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/jackc/pgx/v5"
)

type roleQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func normalizeRoleCodes(roleCodes []string) []string {
	if len(roleCodes) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(roleCodes))
	out := make([]string, 0, len(roleCodes))
	for _, code := range roleCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}

	sort.Strings(out)
	return out
}

func scanRole(row pgx.Row) (*model.Role, error) {
	var role model.Role
	if err := row.Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.CreatedAt); err != nil {
		return nil, err
	}
	return &role, nil
}

func getUserRolesTx(ctx context.Context, q roleQueryer, userID string) ([]*model.Role, error) {
	rows, err := q.Query(ctx, `
		SELECT r.id, r.code, r.name, r.description, r.created_at
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.code ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*model.Role, 0)
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user roles: %w", err)
	}

	return roles, nil
}

func resolveRoleIDsTx(ctx context.Context, tx pgx.Tx, roleCodes []string) (map[string]int64, error) {
	if len(roleCodes) == 0 {
		return map[string]int64{}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT id, code
		FROM roles
		WHERE code = ANY($1::text[])
	`, roleCodes)
	if err != nil {
		return nil, fmt.Errorf("query roles by codes: %w", err)
	}
	defer rows.Close()

	roleIDs := make(map[string]int64, len(roleCodes))
	for rows.Next() {
		var id int64
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, fmt.Errorf("scan role id: %w", err)
		}
		roleIDs[code] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role ids: %w", err)
	}

	if len(roleIDs) != len(roleCodes) {
		return nil, ErrInvalidRoleCode
	}

	return roleIDs, nil
}

func setUserRolesTx(ctx context.Context, tx pgx.Tx, userID string, roleCodes []string, assignedBy *string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear user roles: %w", err)
	}

	if len(roleCodes) == 0 {
		return nil
	}

	roleIDs, err := resolveRoleIDsTx(ctx, tx, roleCodes)
	if err != nil {
		return err
	}

	for _, code := range roleCodes {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
			VALUES ($1, $2, NOW(), $3)
			ON CONFLICT (user_id, role_id) DO UPDATE
			SET assigned_at = EXCLUDED.assigned_at,
			    assigned_by = EXCLUDED.assigned_by
		`, userID, roleIDs[code], assignedBy); err != nil {
			return fmt.Errorf("assign role %q: %w", code, err)
		}
	}

	return nil
}

// ListRoles возвращает все фиксированные роли.
func (r *Repo) ListRoles(ctx context.Context) ([]*model.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, description, created_at
		FROM roles
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*model.Role, 0)
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}

	return roles, nil
}

// GetUserRoles возвращает роли, назначенные пользователю.
func (r *Repo) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	return getUserRolesTx(ctx, r.pool, userID)
}

// SetUserRoles полностью заменяет все роли пользователя.
func (r *Repo) SetUserRoles(ctx context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, error) {
	roleCodes = normalizeRoleCodes(roleCodes)

	var roles []*model.Role
	if err := withTx(ctx, r.pool, func(tx pgx.Tx) error {
		if err := setUserRolesTx(ctx, tx, userID, roleCodes, assignedBy); err != nil {
			return err
		}
		fetched, err := getUserRolesTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		roles = fetched
		return nil
	}); err != nil {
		if errors.Is(err, ErrInvalidRoleCode) {
			return nil, err
		}
		return nil, fmt.Errorf("set user roles: %w", err)
	}

	return roles, nil
}

// AddUserRole добавляет одну роль пользователю.
func (r *Repo) AddUserRole(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil, ErrInvalidRoleCode
	}

	var roles []*model.Role
	if err := withTx(ctx, r.pool, func(tx pgx.Tx) error {
		roleIDs, err := resolveRoleIDsTx(ctx, tx, []string{roleCode})
		if err != nil {
			return err
		}
		roleID := roleIDs[roleCode]

		_, err = tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
			VALUES ($1, $2, NOW(), $3)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, userID, roleID, assignedBy)
		if err != nil {
			return fmt.Errorf("insert user role: %w", err)
		}

		fetched, err := getUserRolesTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		roles = fetched
		return nil
	}); err != nil {
		if errors.Is(err, ErrInvalidRoleCode) {
			return nil, err
		}
		return nil, fmt.Errorf("add user role: %w", err)
	}

	return roles, nil
}

// RemoveUserRole удаляет одну роль у пользователя.
func (r *Repo) RemoveUserRole(ctx context.Context, userID string, roleCode string) ([]*model.Role, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil, ErrInvalidRoleCode
	}

	var roles []*model.Role
	if err := withTx(ctx, r.pool, func(tx pgx.Tx) error {
		roleIDs, err := resolveRoleIDsTx(ctx, tx, []string{roleCode})
		if err != nil {
			return err
		}
		roleID := roleIDs[roleCode]

		_, err = tx.Exec(ctx, `
			DELETE FROM user_roles
			WHERE user_id = $1 AND role_id = $2
		`, userID, roleID)
		if err != nil {
			return fmt.Errorf("delete user role: %w", err)
		}

		fetched, err := getUserRolesTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		roles = fetched
		return nil
	}); err != nil {
		if errors.Is(err, ErrInvalidRoleCode) {
			return nil, err
		}
		return nil, fmt.Errorf("remove user role: %w", err)
	}

	return roles, nil
}

// HasUserWithRole проверяет, есть ли хотя бы один пользователь с указанной ролью.
func (r *Repo) HasUserWithRole(ctx context.Context, roleCode string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE r.code = $1
		)
	`, roleCode).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user role existence: %w", err)
	}
	return exists, nil
}
