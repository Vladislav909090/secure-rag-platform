package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseListRolesReturnsRepoRoles(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser), iamTestRole(RoleSuperAdmin)}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		ListRoles(mock.Anything).
		RunAndReturn(func(context.Context) ([]*model.Role, error) {
			return roles, nil
		})
	uc := newIAMTestUsecase(repo)

	got, err := uc.ListRoles(context.Background())
	require.NoError(t, err)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseGetUserRolesReturnsRolesAndContextVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser)}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSubjectContext(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return iamTestSubject("u1"), nil
		})
	repo.EXPECT().
		GetUserRoles(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)

			return roles, nil
		})
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.GetUserRoles(context.Background(), " u1 ")
	require.NoError(t, err)
	assert.Equal(t, int64(3), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseGetUserRolesRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))

	roles, ctxVer, err := uc.GetUserRoles(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseSetUserRolesNormalizesAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleAccessAdmin), iamTestRole(RoleUser)}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		})
	repo.EXPECT().
		SetUserRoles(mock.Anything, "u1", []string{RoleAccessAdmin, RoleUser}, (*string)(nil)).
		RunAndReturn(func(_ context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, []string{RoleAccessAdmin, RoleUser}, roleCodes)
			assert.Nil(t, assignedBy)

			return roles, nil
		})
	repo.EXPECT().
		IncrementContextVersion(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u1", userID)

			return 4, nil
		})
	repo.EXPECT().
		GetUserRoles(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)

			return roles, nil
		})
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.SetUserRoles(context.Background(), " u1 ", []string{RoleUser, RoleAccessAdmin}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseSetUserRolesRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		})
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.SetUserRoles(context.Background(), "u1", []string{"unknown"}, nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseSetUserRolesReturnsNotFoundWhenUserMissing(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return nil, nil
		})
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.SetUserRoles(context.Background(), "u1", []string{RoleUser}, nil)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseAddUserRoleAddsRoleAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleKnowledgeEditor)}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		})
	repo.EXPECT().
		AddUserRole(mock.Anything, "u1", RoleKnowledgeEditor, (*string)(nil)).
		RunAndReturn(func(_ context.Context, userID string, roleCode string, _ *string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, RoleKnowledgeEditor, roleCode)

			return roles, nil
		})
	repo.EXPECT().
		IncrementContextVersion(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (int64, error) {
			return 5, nil
		})
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.AddUserRole(context.Background(), "u1", RoleKnowledgeEditor, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(5), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseRemoveUserRoleRemovesRoleAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser)}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		})
	repo.EXPECT().
		RemoveUserRole(mock.Anything, "u1", RoleKnowledgeEditor).
		RunAndReturn(func(_ context.Context, userID string, roleCode string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, RoleKnowledgeEditor, roleCode)

			return roles, nil
		})
	repo.EXPECT().
		IncrementContextVersion(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (int64, error) {
			return 6, nil
		})
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.RemoveUserRole(context.Background(), "u1", RoleKnowledgeEditor)
	require.NoError(t, err)
	assert.Equal(t, int64(6), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseAddUserRoleRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		})
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.AddUserRole(context.Background(), "u1", "unknown", nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseRemoveUserRoleRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		})
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.RemoveUserRole(context.Background(), "u1", "unknown")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}
