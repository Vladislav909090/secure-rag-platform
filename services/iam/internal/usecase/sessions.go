package usecase

import (
	"context"
	"strings"

	"secure-rag-platform/services/iam/internal/model"
)

// ListUserSessions возвращает активные сессии текущего или целевого пользователя
func (uc *IAMUsecase) ListUserSessions(
	ctx context.Context,
	principal *Principal,
	userID string,
) ([]*model.UserSession, string, error) {
	if principal == nil {
		return nil, "", ErrUnauthorized
	}

	targetUserID := strings.TrimSpace(userID)
	if targetUserID == "" {
		targetUserID = principal.UserID
	}

	if targetUserID != principal.UserID && !hasAnyRole(principal.Roles, RoleAccessAdmin, RoleSuperAdmin) {
		return nil, "", ErrForbidden
	}

	if err := uc.ensureUserExists(ctx, targetUserID); err != nil {
		return nil, "", err
	}

	sessions, err := uc.repo.ListActiveUserSessions(ctx, targetUserID)
	if err != nil {
		return nil, "", err
	}

	return sessions, targetUserID, nil
}

// RevokeSession отзывает одну сессию с проверкой прав доступа
func (uc *IAMUsecase) RevokeSession(ctx context.Context, principal *Principal, sessionID string) (bool, error) {
	return uc.Logout(ctx, principal, sessionID)
}

// RevokeAllUserSessions отзывает все активные сессии
func (uc *IAMUsecase) RevokeAllUserSessions(
	ctx context.Context,
	principal *Principal,
	userID string,
) (*LogoutAllResult, error) {
	return uc.LogoutAll(ctx, principal, userID)
}
