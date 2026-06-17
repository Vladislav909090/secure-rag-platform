package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

// ListUsers возвращает всех пользователей с ролями и атрибутами
func (uc *IAMUsecase) ListUsers(ctx context.Context) ([]*model.UserView, error) {
	return uc.repo.ListUserViews(ctx)
}

// GetUser возвращает одного пользователя с ролями и атрибутами
func (uc *IAMUsecase) GetUser(ctx context.Context, userID string) (*model.UserView, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}

	view, err := uc.repo.GetUserView(ctx, userID)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, ErrNotFound
	}

	return view, nil
}

// CreateUser создает пользователя, атрибуты и назначения ролей
func (uc *IAMUsecase) CreateUser(ctx context.Context, input CreateUserInput) (*model.UserView, error) {
	login := strings.TrimSpace(input.Login)
	if login == "" {
		return nil, ErrInvalidArgument
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidArgument
	}

	roleCodes, err := normalizeRoleCodes(input.RoleCodes)
	if err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	created, err := uc.repo.CreateUser(ctx, repository.CreateUserInput{
		Login:        login,
		PasswordHash: passwordHash,
		IsActive:     isActive,
		RoleCodes:    roleCodes,
		Attributes:   input.Attributes,
		CreatedBy:    input.CreatedBy,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}
		if errors.Is(err, repository.ErrInvalidRoleCode) {
			return nil, ErrInvalidArgument
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	view, err := uc.repo.GetUserView(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, ErrNotFound
	}

	uc.InvalidateSubjectContextCache(ctx, created.ID)

	return view, nil
}

// UpdateUser обновляет поля пользователя и увеличивает версию контекста
func (uc *IAMUsecase) UpdateUser(ctx context.Context, input UpdateUserInput) (*model.UserView, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	if input.UserID == "" {
		return nil, ErrInvalidArgument
	}

	if input.Login != nil {
		trimmed := strings.TrimSpace(*input.Login)
		if trimmed == "" {
			return nil, ErrInvalidArgument
		}
		input.Login = &trimmed
	}

	var passwordHash *string
	if input.Password != nil {
		h, err := hashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &h
	}

	if input.Login == nil && passwordHash == nil && input.IsActive == nil {
		return nil, ErrInvalidArgument
	}

	_, err := uc.repo.UpdateUser(ctx, repository.UpdateUserInput{
		UserID:       input.UserID,
		Login:        input.Login,
		PasswordHash: passwordHash,
		IsActive:     input.IsActive,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}

		return nil, err
	}

	_, err = uc.bumpContextVersion(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	view, err := uc.repo.GetUserView(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, ErrNotFound
	}

	return view, nil
}
