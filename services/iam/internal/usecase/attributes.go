package usecase

import (
	"context"
	"strings"
)

// GetUserAttributes возвращает атрибуты пользователя и текущую версию контекста
func (uc *IAMUsecase) GetUserAttributes(ctx context.Context, userID string) (map[string]any, int64, error) {
	subject, err := uc.GetSubjectContext(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, 0, err
	}

	return subject.Attributes, subject.CtxVer, nil
}

// ReplaceUserAttributes полностью заменяет атрибуты и увеличивает версию контекста
func (uc *IAMUsecase) ReplaceUserAttributes(
	ctx context.Context,
	userID string,
	attrs map[string]any,
	updatedBy *string,
) (map[string]any, int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	updated, err := uc.repo.ReplaceUserAttributes(ctx, userID, attrs, updatedBy)
	if err != nil {
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return updated, ctxVer, nil
}

// PatchUserAttributes обновляет атрибуты верхнего уровня и версию контекста
func (uc *IAMUsecase) PatchUserAttributes(
	ctx context.Context,
	userID string,
	patch map[string]any,
	updatedBy *string,
) (map[string]any, int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	current, err := uc.repo.GetUserAttributes(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	merged := mergeAttributes(current, patch)
	updated, err := uc.repo.ReplaceUserAttributes(ctx, userID, merged, updatedBy)
	if err != nil {
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return updated, ctxVer, nil
}

// DeleteUserAttributeKey удаляет один ключ и увеличивает версию контекста
func (uc *IAMUsecase) DeleteUserAttributeKey(
	ctx context.Context,
	userID string,
	key string,
	updatedBy *string,
) (map[string]any, int64, error) {
	userID = strings.TrimSpace(userID)
	key = strings.TrimSpace(key)
	if userID == "" || key == "" {
		return nil, 0, ErrInvalidArgument
	}
	if err := uc.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	updated, err := uc.repo.DeleteUserAttributeKey(ctx, userID, key, updatedBy)
	if err != nil {
		return nil, 0, err
	}

	ctxVer, err := uc.bumpContextVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return updated, ctxVer, nil
}
