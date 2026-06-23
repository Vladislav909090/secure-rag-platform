package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseGetUserAttributesReturnsSubjectAttributes(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return iamTestSubject("u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	attrs, ctxVer, err := uc.GetUserAttributes(context.Background(), " u1 ")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"department": "search"}, attrs)
	assert.Equal(t, int64(3), ctxVer)
}

func TestIAMUsecaseReplaceUserAttributesReplacesAndBumpsVersion(t *testing.T) {
	t.Parallel()

	updated := map[string]any{"team": "rag"}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		},
		replaceUserAttributes: func(_ context.Context, userID string, attrs map[string]any, _ *string) (map[string]any, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, updated, attrs)

			return updated, nil
		},
		incrementContextVersion: func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u1", userID)

			return 4, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	attrs, ctxVer, err := uc.ReplaceUserAttributes(context.Background(), " u1 ", updated, nil)
	require.NoError(t, err)
	assert.Equal(t, updated, attrs)
	assert.Equal(t, int64(4), ctxVer)
}

func TestIAMUsecaseReplaceUserAttributesRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	attrs, ctxVer, err := uc.ReplaceUserAttributes(context.Background(), " ", map[string]any{}, nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, attrs)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecasePatchUserAttributesMergesAndBumpsVersion(t *testing.T) {
	t.Parallel()

	updated := map[string]any{"department": "search", "team": "rag"}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
		getUserAttributes: func(_ context.Context, userID string) (map[string]any, error) {
			assert.Equal(t, "u1", userID)

			return map[string]any{"department": "old"}, nil
		},
		replaceUserAttributes: func(_ context.Context, userID string, attrs map[string]any, _ *string) (map[string]any, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, updated, attrs)

			return updated, nil
		},
		incrementContextVersion: func(context.Context, string) (int64, error) {
			return 5, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	attrs, ctxVer, err := uc.PatchUserAttributes(context.Background(), "u1", updated, nil)
	require.NoError(t, err)
	assert.Equal(t, updated, attrs)
	assert.Equal(t, int64(5), ctxVer)
}

func TestIAMUsecasePatchUserAttributesRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	attrs, ctxVer, err := uc.PatchUserAttributes(context.Background(), " ", map[string]any{}, nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, attrs)
	assert.Zero(t, ctxVer)
}

func TestIAMUsecaseDeleteUserAttributeKeyDeletesAndBumpsVersion(t *testing.T) {
	t.Parallel()

	updated := map[string]any{"department": "search"}
	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
		deleteUserAttributeKey: func(_ context.Context, userID string, key string, _ *string) (map[string]any, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "team", key)

			return updated, nil
		},
		incrementContextVersion: func(context.Context, string) (int64, error) {
			return 6, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	attrs, ctxVer, err := uc.DeleteUserAttributeKey(context.Background(), " u1 ", " team ", nil)
	require.NoError(t, err)
	assert.Equal(t, updated, attrs)
	assert.Equal(t, int64(6), ctxVer)
}

func TestIAMUsecaseDeleteUserAttributeKeyRejectsEmptyKey(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	attrs, ctxVer, err := uc.DeleteUserAttributeKey(context.Background(), "u1", " ", nil)
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, attrs)
	assert.Zero(t, ctxVer)
}
