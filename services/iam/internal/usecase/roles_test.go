package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseListRolesReturnsRepoRoles(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser), iamTestRole(RoleSuperAdmin)}
	repo := &mockIAMRepo{
		t: t,
		listRoles: func(context.Context) ([]*model.Role, error) {
			return roles, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.ListRoles(context.Background())
	require.NoError(t, err)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseGetUserRolesReturnsRolesAndContextVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser)}
	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return iamTestSubject("u1"), nil
		},
		getUserRoles: func(_ context.Context, userID string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)

			return roles, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.GetUserRoles(context.Background(), " u1 ")
	require.NoError(t, err)
	assert.Equal(t, int64(3), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseGetUserRolesRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	roles, ctxVer, err := uc.GetUserRoles(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseSetUserRolesNormalizesAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleAccessAdmin), iamTestRole(RoleUser)}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		},
		setUserRoles: func(_ context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, []string{RoleAccessAdmin, RoleUser}, roleCodes)
			assert.Nil(t, assignedBy)

			return roles, nil
		},
		incrementContextVersion: func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u1", userID)

			return 4, nil
		},
		getUserRoles: func(_ context.Context, userID string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)

			return roles, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.SetUserRoles(context.Background(), " u1 ", []string{RoleUser, RoleAccessAdmin}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseSetUserRolesRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.SetUserRoles(context.Background(), "u1", []string{"unknown"}, nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseSetUserRolesReturnsNotFoundWhenUserMissing(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return nil, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.SetUserRoles(context.Background(), "u1", []string{RoleUser}, nil)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseAddUserRoleAddsRoleAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleKnowledgeEditor)}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
		addUserRole: func(_ context.Context, userID string, roleCode string, _ *string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, RoleKnowledgeEditor, roleCode)

			return roles, nil
		},
		incrementContextVersion: func(context.Context, string) (int64, error) {
			return 5, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.AddUserRole(context.Background(), "u1", RoleKnowledgeEditor, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(5), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseRemoveUserRoleRemovesRoleAndBumpsVersion(t *testing.T) {
	t.Parallel()

	roles := []*model.Role{iamTestRole(RoleUser)}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
		removeUserRole: func(_ context.Context, userID string, roleCode string) ([]*model.Role, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, RoleKnowledgeEditor, roleCode)

			return roles, nil
		},
		incrementContextVersion: func(context.Context, string) (int64, error) {
			return 6, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, ctxVer, err := uc.RemoveUserRole(context.Background(), "u1", RoleKnowledgeEditor)
	require.NoError(t, err)
	assert.Equal(t, int64(6), ctxVer)
	assert.Equal(t, roles, got)
}

func TestIAMUsecaseAddUserRoleRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.AddUserRole(context.Background(), "u1", "unknown", nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseRemoveUserRoleRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	roles, ctxVer, err := uc.RemoveUserRole(context.Background(), "u1", "unknown")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, roles)
	assert.Zero(t, ctxVer)
}
