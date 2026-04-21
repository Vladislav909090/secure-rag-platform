package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"
)

// Login выполняет аутентификацию по логину/паролю и выдает пару токенов доступа и обновления.
func (uc *IAMUsecase) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	login := strings.TrimSpace(input.Login)
	password := input.Password
	if login == "" || password == "" {
		return nil, ErrInvalidArgument
	}

	rateKey := loginRateKeyPrefix + login
	if err := uc.checkRateLimit(ctx, rateKey, uc.cfg.LoginRateLimit, uc.cfg.LoginRateWindow); err != nil {
		return nil, err
	}

	user, err := uc.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil || !checkPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	subject, err := uc.GetSubjectContext(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := randomToken(48)
	if err != nil {
		return nil, err
	}
	refreshHash := hashOpaqueToken(refreshToken)

	expiresAt := time.Now().UTC().Add(uc.cfg.RefreshTokenTTL)
	session, err := uc.repo.CreateSession(ctx, repository.CreateSessionInput{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	accessToken, expiresIn, _, err := uc.issueAccessToken(subject, session.ID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    tokenTypeBearer,
	}, nil
}

// RefreshToken выполняет ротацию токена обновления и возвращает новую пару токенов доступа и обновления.
func (uc *IAMUsecase) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenPair, error) {
	refreshRaw := strings.TrimSpace(input.RefreshToken)
	if refreshRaw == "" {
		return nil, ErrInvalidArgument
	}

	rateKey := refreshRateKeyPrefix + hashOpaqueToken(refreshRaw)
	if err := uc.checkRateLimit(ctx, rateKey, uc.cfg.RefreshRateLimit, uc.cfg.RefreshRateWindow); err != nil {
		return nil, err
	}

	refreshHash := hashOpaqueToken(refreshRaw)
	session, err := uc.repo.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidCredentials
	}
	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrSessionExpired
	}

	user, err := uc.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	subject, err := uc.GetSubjectContext(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := randomToken(48)
	if err != nil {
		return nil, err
	}
	newRefreshHash := hashOpaqueToken(newRefreshToken)
	newExpiresAt := time.Now().UTC().Add(uc.cfg.RefreshTokenTTL)

	_, err = uc.repo.RotateSessionRefreshHash(
		ctx,
		session.ID,
		newRefreshHash,
		newExpiresAt,
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionRevoked
		}
		return nil, err
	}

	accessToken, expiresIn, _, err := uc.issueAccessToken(subject, session.ID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    tokenTypeBearer,
	}, nil
}

// Logout отзывает одну сессию.
func (uc *IAMUsecase) Logout(ctx context.Context, principal *Principal, sessionID string) (bool, error) {
	if principal == nil {
		return false, ErrUnauthorized
	}

	targetSessionID := strings.TrimSpace(sessionID)
	if targetSessionID == "" {
		targetSessionID = principal.SessionID
	}

	if targetSessionID != principal.SessionID && !hasAnyRole(principal.Roles, RoleAccessAdmin, RoleSuperAdmin) {
		return false, ErrForbidden
	}

	targetSession, err := uc.repo.GetSessionByID(ctx, targetSessionID)
	if err != nil {
		return false, err
	}
	if targetSession == nil {
		return false, ErrNotFound
	}

	if targetSession.UserID != principal.UserID && !hasAnyRole(principal.Roles, RoleAccessAdmin, RoleSuperAdmin) {
		return false, ErrForbidden
	}

	revoked, err := uc.repo.RevokeSession(ctx, targetSessionID)
	if err != nil {
		return false, err
	}
	return revoked, nil
}

// LogoutAll отзывает все активные сессии целевого пользователя.
func (uc *IAMUsecase) LogoutAll(ctx context.Context, principal *Principal, userID string) (*LogoutAllResult, error) {
	if principal == nil {
		return nil, ErrUnauthorized
	}

	targetUserID := strings.TrimSpace(userID)
	if targetUserID == "" {
		targetUserID = principal.UserID
	}

	if targetUserID != principal.UserID && !hasAnyRole(principal.Roles, RoleAccessAdmin, RoleSuperAdmin) {
		return nil, ErrForbidden
	}

	targetUser, err := uc.repo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, ErrNotFound
	}

	revokedCount, err := uc.repo.RevokeAllUserSessions(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	return &LogoutAllResult{
		UserID:       targetUserID,
		RevokedCount: revokedCount,
		CtxVer:       targetUser.CtxVer,
	}, nil
}

// GetMe возвращает текущий нормализованный контекст субъекта.
func (uc *IAMUsecase) GetMe(ctx context.Context, principal *Principal) (*model.SubjectContext, error) {
	if principal == nil {
		return nil, ErrUnauthorized
	}
	return uc.GetSubjectContext(ctx, principal.UserID)
}
