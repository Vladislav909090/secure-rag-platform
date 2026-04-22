package repository

import (
	"context"
	"fmt"

	"secure-rag-platform/services/iam/internal/model"
)

func rolesToCodes(roles []*model.Role) []string {
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	return codes
}

// GetSubjectContext возвращает нормализованный контекст безопасности пользователя.
func (r *Repo) GetSubjectContext(ctx context.Context, userID string) (*model.SubjectContext, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	attrs, err := r.GetUserAttributes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user attributes: %w", err)
	}

	return &model.SubjectContext{
		UserID:     user.ID,
		Login:      user.Login,
		IsActive:   user.IsActive,
		Roles:      rolesToCodes(roles),
		Attributes: normalizeAttributes(attrs),
		CtxVer:     user.CtxVer,
	}, nil
}

// GetUserView возвращает пользователя вместе с ролями и атрибутами.
func (r *Repo) GetUserView(ctx context.Context, userID string) (*model.UserView, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	attrs, err := r.GetUserAttributes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user attributes: %w", err)
	}

	return &model.UserView{
		ID:         user.ID,
		Login:      user.Login,
		IsActive:   user.IsActive,
		CtxVer:     user.CtxVer,
		Roles:      rolesToCodes(roles),
		Attributes: normalizeAttributes(attrs),
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

// ListUserViews возвращает всех пользователей с ролями и атрибутами.
func (r *Repo) ListUserViews(ctx context.Context) ([]*model.UserView, error) {
	users, err := r.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]*model.UserView, 0, len(users))
	for _, user := range users {
		roles, err := r.GetUserRoles(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("get user %s roles: %w", user.ID, err)
		}

		attrs, err := r.GetUserAttributes(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("get user %s attributes: %w", user.ID, err)
		}

		views = append(views, &model.UserView{
			ID:         user.ID,
			Login:      user.Login,
			IsActive:   user.IsActive,
			CtxVer:     user.CtxVer,
			Roles:      rolesToCodes(roles),
			Attributes: normalizeAttributes(attrs),
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		})
	}

	return views, nil
}
