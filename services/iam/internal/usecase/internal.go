package usecase

import (
	"context"
	"errors"
	"strings"

	"secure-rag-platform/services/iam/internal/model"
)

func isValidationFailure(err error) bool {
	return errors.Is(err, ErrInvalidArgument) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInactiveUser) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrSessionRevoked) ||
		errors.Is(err, ErrNotFound)
}

// ValidateAccessToken проверяет токен и при успешной валидации возвращает контекст субъекта
func (uc *IAMUsecase) ValidateAccessToken(ctx context.Context, accessToken string) (*ValidateTokenResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return &ValidateTokenResult{Valid: false, Reason: ErrInvalidArgument.Error()}, nil
	}

	principal, subject, err := uc.AuthenticateAccessToken(ctx, accessToken)
	if err != nil {
		if isValidationFailure(err) {
			return &ValidateTokenResult{Valid: false, Reason: err.Error()}, nil
		}

		return nil, err
	}

	return &ValidateTokenResult{
		Valid:         true,
		Reason:        "",
		Subject:       subject,
		Principal:     principal,
		ExpiresAtUnix: principal.ExpiresAt.Unix(),
	}, nil
}

// GetSubjectContextByUserID возвращает нормализованный контекст для внутренних вызовов
func (uc *IAMUsecase) GetSubjectContextByUserID(ctx context.Context, userID string) (*model.SubjectContext, error) {
	return uc.GetSubjectContext(ctx, userID)
}
