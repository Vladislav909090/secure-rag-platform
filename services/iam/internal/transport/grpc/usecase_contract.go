package grpc

import (
	"context"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"
)

type IAMUsecaseContract interface {
	AddUserRole(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, int64, error)
	AuthenticateAccessToken(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error)
	CreateUser(ctx context.Context, input usecase.CreateUserInput) (*model.UserView, error)
	DeleteUserAttributeKey(ctx context.Context, userID string, key string, updatedBy *string) (map[string]any, int64, error)
	GetMe(ctx context.Context, principal *usecase.Principal) (*model.SubjectContext, error)
	GetSubjectContextByUserID(ctx context.Context, userID string) (*model.SubjectContext, error)
	GetUser(ctx context.Context, userID string) (*model.UserView, error)
	GetUserAttributes(ctx context.Context, userID string) (map[string]any, int64, error)
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, int64, error)
	ListRoles(ctx context.Context) ([]*model.Role, error)
	ListUserSessions(ctx context.Context, principal *usecase.Principal, userID string) ([]*model.UserSession, string, error)
	ListUsers(ctx context.Context) ([]*model.UserView, error)
	Login(ctx context.Context, input usecase.LoginInput) (*usecase.TokenPair, error)
	Logout(ctx context.Context, principal *usecase.Principal, sessionID string) (bool, error)
	LogoutAll(ctx context.Context, principal *usecase.Principal, userID string) (*usecase.LogoutAllResult, error)
	PatchUserAttributes(ctx context.Context, userID string, patch map[string]any, updatedBy *string) (map[string]any, int64, error)
	RefreshToken(ctx context.Context, input usecase.RefreshTokenInput) (*usecase.TokenPair, error)
	RemoveUserRole(ctx context.Context, userID string, roleCode string) ([]*model.Role, int64, error)
	ReplaceUserAttributes(ctx context.Context, userID string, attrs map[string]any, updatedBy *string) (map[string]any, int64, error)
	RevokeAllUserSessions(ctx context.Context, principal *usecase.Principal, userID string) (*usecase.LogoutAllResult, error)
	RevokeSession(ctx context.Context, principal *usecase.Principal, sessionID string) (bool, error)
	SetUserRoles(ctx context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, int64, error)
	UpdateUser(ctx context.Context, input usecase.UpdateUserInput) (*model.UserView, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*usecase.ValidateTokenResult, error)
}

func usecaseOrNil(uc *usecase.IAMUsecase) IAMUsecaseContract {
	if uc == nil {
		return nil
	}

	return uc
}
