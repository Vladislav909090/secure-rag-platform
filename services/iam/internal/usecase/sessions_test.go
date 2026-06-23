package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseListUserSessionsDefaultsToPrincipalUser(t *testing.T) {
	t.Parallel()

	sessions := []*model.UserSession{iamTestSession("s1", "u1")}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		})
	repo.EXPECT().
		ListActiveUserSessions(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) ([]*model.UserSession, error) {
			assert.Equal(t, "u1", userID)

			return sessions, nil
		})
	uc := newIAMTestUsecase(repo)

	got, targetUserID, err := uc.ListUserSessions(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleUser}}, "")
	require.NoError(t, err)
	assert.Equal(t, "u1", targetUserID)
	assert.Equal(t, sessions, got)
}

func TestIAMUsecaseListUserSessionsAllowsAdminTargetUser(t *testing.T) {
	t.Parallel()

	sessions := []*model.UserSession{iamTestSession("s1", "u2")}
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u2").
		RunAndReturn(func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u2", userID)

			return iamTestUser("u2"), nil
		})
	repo.EXPECT().
		ListActiveUserSessions(mock.Anything, "u2").
		RunAndReturn(func(_ context.Context, userID string) ([]*model.UserSession, error) {
			assert.Equal(t, "u2", userID)

			return sessions, nil
		})
	uc := newIAMTestUsecase(repo)

	got, targetUserID, err := uc.ListUserSessions(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleAccessAdmin}}, " u2 ")
	require.NoError(t, err)
	assert.Equal(t, "u2", targetUserID)
	assert.Equal(t, sessions, got)
}

func TestIAMUsecaseListUserSessionsRejectsOtherUserForNonAdmin(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))

	got, targetUserID, err := uc.ListUserSessions(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleUser}}, "u2")
	require.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, got)
	assert.Empty(t, targetUserID)
}

func TestIAMUsecaseRevokeSessionDelegatesToLogout(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSessionByID(mock.Anything, "s1").
		RunAndReturn(func(_ context.Context, sessionID string) (*model.UserSession, error) {
			assert.Equal(t, "s1", sessionID)

			return iamTestSession("s1", "u1"), nil
		})
	repo.EXPECT().
		RevokeSession(mock.Anything, "s1").
		RunAndReturn(func(_ context.Context, sessionID string) (bool, error) {
			assert.Equal(t, "s1", sessionID)

			return true, nil
		})
	uc := newIAMTestUsecase(repo)

	revoked, err := uc.RevokeSession(context.Background(), &Principal{UserID: "u1", SessionID: "s1"}, "")
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestIAMUsecaseRevokeAllUserSessionsDelegatesToLogoutAll(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetUserByID(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		})
	repo.EXPECT().
		RevokeAllUserSessions(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u1", userID)

			return 2, nil
		})
	uc := newIAMTestUsecase(repo)

	got, err := uc.RevokeAllUserSessions(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleUser}}, "")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u1", got.UserID)
	assert.Equal(t, int64(2), got.RevokedCount)
	assert.Equal(t, int64(3), got.CtxVer)
}
