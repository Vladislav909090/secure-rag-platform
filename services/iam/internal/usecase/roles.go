package usecase

import (
	"context"
	"errors"
	"strings"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"
)

func (uc *IAMUsecase) ensureUserExists(ctx context.Context, userID string) error {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotFound
	}
	return nil
}

// ListRoles возвращает фиксированные системные роли.
func (uc *IAMUsecase) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return uc.repo.ListRoles(ctx)
}

// GetUserRoles возвращает назначенные роли и текущую версию контекста.
func (uc *IAMUsecase) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, ErrInvalidArgument
	}

	subject, err := uc.GetSubjectContext(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, err
	}

	roles, err := uc.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return roles, subject.CtxVer, nil
}

// SetUserRoles полностью заменяет набор ролей и увеличивает версию контекста.
func (uc *IAMUsecase) SetUserRoles(ctx context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	normalized, err := normalizeRoleCodes(roleCodes)
	if err != nil {
		return nil, 0, err
	}

	_, err = uc.repo.SetUserRoles(ctx, userID, normalized, assignedBy)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidRoleCode) {
			return nil, 0, ErrInvalidArgument
		}
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	roles, err := uc.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return roles, ctxVer, nil
}

// AddUserRole добавляет роль и увеличивает версию контекста.
func (uc *IAMUsecase) AddUserRole(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, int64, error) {
	userID = strings.TrimSpace(userID)
	roleCode = strings.TrimSpace(roleCode)
	if userID == "" || roleCode == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}
	if _, ok := fixedRoleSet[roleCode]; !ok {
		return nil, 0, ErrInvalidArgument
	}

	roles, err := uc.repo.AddUserRole(ctx, userID, roleCode, assignedBy)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidRoleCode) {
			return nil, 0, ErrInvalidArgument
		}
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return roles, ctxVer, nil
}

// RemoveUserRole удаляет роль и увеличивает версию контекста.
func (uc *IAMUsecase) RemoveUserRole(ctx context.Context, userID string, roleCode string) ([]*model.Role, int64, error) {
	userID = strings.TrimSpace(userID)
	roleCode = strings.TrimSpace(roleCode)
	if userID == "" || roleCode == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}
	if _, ok := fixedRoleSet[roleCode]; !ok {
		return nil, 0, ErrInvalidArgument
	}

	roles, err := uc.repo.RemoveUserRole(ctx, userID, roleCode)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidRoleCode) {
			return nil, 0, ErrInvalidArgument
		}
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return roles, ctxVer, nil
}
