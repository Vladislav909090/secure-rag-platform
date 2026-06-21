package grpc

import (
	"context"
	"fmt"
	"testing"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

type mockIAMUsecase struct {
	t *testing.T

	addUserRole               func(context.Context, string, string, *string) ([]*model.Role, int64, error)
	authenticateAccessToken   func(context.Context, string) (*usecase.Principal, *model.SubjectContext, error)
	createUser                func(context.Context, usecase.CreateUserInput) (*model.UserView, error)
	deleteUserAttributeKey    func(context.Context, string, string, *string) (map[string]any, int64, error)
	getMe                     func(context.Context, *usecase.Principal) (*model.SubjectContext, error)
	getSubjectContextByUserID func(context.Context, string) (*model.SubjectContext, error)
	getUser                   func(context.Context, string) (*model.UserView, error)
	getUserAttributes         func(context.Context, string) (map[string]any, int64, error)
	getUserRoles              func(context.Context, string) ([]*model.Role, int64, error)
	listRoles                 func(context.Context) ([]*model.Role, error)
	listUserSessions          func(context.Context, *usecase.Principal, string) ([]*model.UserSession, string, error)
	listUsers                 func(context.Context) ([]*model.UserView, error)
	login                     func(context.Context, usecase.LoginInput) (*usecase.TokenPair, error)
	logout                    func(context.Context, *usecase.Principal, string) (bool, error)
	logoutAll                 func(context.Context, *usecase.Principal, string) (*usecase.LogoutAllResult, error)
	patchUserAttributes       func(context.Context, string, map[string]any, *string) (map[string]any, int64, error)
	refreshToken              func(context.Context, usecase.RefreshTokenInput) (*usecase.TokenPair, error)
	removeUserRole            func(context.Context, string, string) ([]*model.Role, int64, error)
	replaceUserAttributes     func(context.Context, string, map[string]any, *string) (map[string]any, int64, error)
	revokeAllUserSessions     func(context.Context, *usecase.Principal, string) (*usecase.LogoutAllResult, error)
	revokeSession             func(context.Context, *usecase.Principal, string) (bool, error)
	setUserRoles              func(context.Context, string, []string, *string) ([]*model.Role, int64, error)
	updateUser                func(context.Context, usecase.UpdateUserInput) (*model.UserView, error)
	validateAccessToken       func(context.Context, string) (*usecase.ValidateTokenResult, error)
}

var _ iamUsecase = (*mockIAMUsecase)(nil)

func (m *mockIAMUsecase) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected usecase call", "unexpected usecase call: %s", name)
	}
	panic(fmt.Sprintf("unexpected usecase call: %s", name))
}

func (m *mockIAMUsecase) AddUserRole(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, int64, error) {
	if m.addUserRole == nil {
		m.unexpected("AddUserRole")
	}

	return m.addUserRole(ctx, userID, roleCode, assignedBy)
}

func (m *mockIAMUsecase) AuthenticateAccessToken(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
	if m.authenticateAccessToken == nil {
		m.unexpected("AuthenticateAccessToken")
	}

	return m.authenticateAccessToken(ctx, accessToken)
}

func (m *mockIAMUsecase) CreateUser(ctx context.Context, input usecase.CreateUserInput) (*model.UserView, error) {
	if m.createUser == nil {
		m.unexpected("CreateUser")
	}

	return m.createUser(ctx, input)
}

func (m *mockIAMUsecase) DeleteUserAttributeKey(ctx context.Context, userID string, key string, updatedBy *string) (map[string]any, int64, error) {
	if m.deleteUserAttributeKey == nil {
		m.unexpected("DeleteUserAttributeKey")
	}

	return m.deleteUserAttributeKey(ctx, userID, key, updatedBy)
}

func (m *mockIAMUsecase) GetMe(ctx context.Context, principal *usecase.Principal) (*model.SubjectContext, error) {
	if m.getMe == nil {
		m.unexpected("GetMe")
	}

	return m.getMe(ctx, principal)
}

func (m *mockIAMUsecase) GetSubjectContextByUserID(ctx context.Context, userID string) (*model.SubjectContext, error) {
	if m.getSubjectContextByUserID == nil {
		m.unexpected("GetSubjectContextByUserID")
	}

	return m.getSubjectContextByUserID(ctx, userID)
}

func (m *mockIAMUsecase) GetUser(ctx context.Context, userID string) (*model.UserView, error) {
	if m.getUser == nil {
		m.unexpected("GetUser")
	}

	return m.getUser(ctx, userID)
}

func (m *mockIAMUsecase) GetUserAttributes(ctx context.Context, userID string) (map[string]any, int64, error) {
	if m.getUserAttributes == nil {
		m.unexpected("GetUserAttributes")
	}

	return m.getUserAttributes(ctx, userID)
}

func (m *mockIAMUsecase) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, int64, error) {
	if m.getUserRoles == nil {
		m.unexpected("GetUserRoles")
	}

	return m.getUserRoles(ctx, userID)
}

func (m *mockIAMUsecase) ListRoles(ctx context.Context) ([]*model.Role, error) {
	if m.listRoles == nil {
		m.unexpected("ListRoles")
	}

	return m.listRoles(ctx)
}

func (m *mockIAMUsecase) ListUserSessions(ctx context.Context, principal *usecase.Principal, userID string) ([]*model.UserSession, string, error) {
	if m.listUserSessions == nil {
		m.unexpected("ListUserSessions")
	}

	return m.listUserSessions(ctx, principal, userID)
}

func (m *mockIAMUsecase) ListUsers(ctx context.Context) ([]*model.UserView, error) {
	if m.listUsers == nil {
		m.unexpected("ListUsers")
	}

	return m.listUsers(ctx)
}

func (m *mockIAMUsecase) Login(ctx context.Context, input usecase.LoginInput) (*usecase.TokenPair, error) {
	if m.login == nil {
		m.unexpected("Login")
	}

	return m.login(ctx, input)
}

func (m *mockIAMUsecase) Logout(ctx context.Context, principal *usecase.Principal, sessionID string) (bool, error) {
	if m.logout == nil {
		m.unexpected("Logout")
	}

	return m.logout(ctx, principal, sessionID)
}

func (m *mockIAMUsecase) LogoutAll(ctx context.Context, principal *usecase.Principal, userID string) (*usecase.LogoutAllResult, error) {
	if m.logoutAll == nil {
		m.unexpected("LogoutAll")
	}

	return m.logoutAll(ctx, principal, userID)
}

func (m *mockIAMUsecase) PatchUserAttributes(ctx context.Context, userID string, patch map[string]any, updatedBy *string) (map[string]any, int64, error) {
	if m.patchUserAttributes == nil {
		m.unexpected("PatchUserAttributes")
	}

	return m.patchUserAttributes(ctx, userID, patch, updatedBy)
}

func (m *mockIAMUsecase) RefreshToken(ctx context.Context, input usecase.RefreshTokenInput) (*usecase.TokenPair, error) {
	if m.refreshToken == nil {
		m.unexpected("RefreshToken")
	}

	return m.refreshToken(ctx, input)
}

func (m *mockIAMUsecase) RemoveUserRole(ctx context.Context, userID string, roleCode string) ([]*model.Role, int64, error) {
	if m.removeUserRole == nil {
		m.unexpected("RemoveUserRole")
	}

	return m.removeUserRole(ctx, userID, roleCode)
}

func (m *mockIAMUsecase) ReplaceUserAttributes(ctx context.Context, userID string, attrs map[string]any, updatedBy *string) (map[string]any, int64, error) {
	if m.replaceUserAttributes == nil {
		m.unexpected("ReplaceUserAttributes")
	}

	return m.replaceUserAttributes(ctx, userID, attrs, updatedBy)
}

func (m *mockIAMUsecase) RevokeAllUserSessions(ctx context.Context, principal *usecase.Principal, userID string) (*usecase.LogoutAllResult, error) {
	if m.revokeAllUserSessions == nil {
		m.unexpected("RevokeAllUserSessions")
	}

	return m.revokeAllUserSessions(ctx, principal, userID)
}

func (m *mockIAMUsecase) RevokeSession(ctx context.Context, principal *usecase.Principal, sessionID string) (bool, error) {
	if m.revokeSession == nil {
		m.unexpected("RevokeSession")
	}

	return m.revokeSession(ctx, principal, sessionID)
}

func (m *mockIAMUsecase) SetUserRoles(ctx context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, int64, error) {
	if m.setUserRoles == nil {
		m.unexpected("SetUserRoles")
	}

	return m.setUserRoles(ctx, userID, roleCodes, assignedBy)
}

func (m *mockIAMUsecase) UpdateUser(ctx context.Context, input usecase.UpdateUserInput) (*model.UserView, error) {
	if m.updateUser == nil {
		m.unexpected("UpdateUser")
	}

	return m.updateUser(ctx, input)
}

func (m *mockIAMUsecase) ValidateAccessToken(ctx context.Context, accessToken string) (*usecase.ValidateTokenResult, error) {
	if m.validateAccessToken == nil {
		m.unexpected("ValidateAccessToken")
	}

	return m.validateAccessToken(ctx, accessToken)
}

func authContext(token string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}
