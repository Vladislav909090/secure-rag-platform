package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseValidateAccessTokenReturnsValidResult(t *testing.T) {
	t.Parallel()

	subject := iamTestSubject("u1")
	session := iamTestSession("s1", "u1")
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSubjectContext(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return subject, nil
		})
	repo.EXPECT().
		GetSessionByID(mock.Anything, "s1").
		RunAndReturn(func(_ context.Context, sessionID string) (*model.UserSession, error) {
			assert.Equal(t, "s1", sessionID)

			return session, nil
		})
	uc := newIAMTestUsecase(repo)
	token, _, _, err := uc.issueAccessToken(subject, "s1")
	require.NoError(t, err)

	got, err := uc.ValidateAccessToken(context.Background(), token)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.Valid)
	assert.Empty(t, got.Reason)
	assert.Same(t, subject, got.Subject)
	require.NotNil(t, got.Principal)
	assert.Equal(t, "u1", got.Principal.UserID)
	assert.Equal(t, "s1", got.Principal.SessionID)
	assert.Positive(t, got.ExpiresAtUnix)
}

func TestIAMUsecaseValidateAccessTokenReturnsInvalidResult(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))

	got, err := uc.ValidateAccessToken(context.Background(), " ")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Valid)
	assert.Equal(t, ErrInvalidArgument.Error(), got.Reason)
}

func TestIAMUsecaseValidateAccessTokenMapsMalformedTokenToInvalidResult(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))

	got, err := uc.ValidateAccessToken(context.Background(), "not-a-jwt")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Valid)
	assert.Equal(t, ErrUnauthorized.Error(), got.Reason)
}

func TestIAMUsecaseGetSubjectContextByUserIDDelegatesToSubjectContext(t *testing.T) {
	t.Parallel()

	subject := iamTestSubject("u1")
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSubjectContext(mock.Anything, "u1").
		RunAndReturn(func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return subject, nil
		})
	uc := newIAMTestUsecase(repo)

	got, err := uc.GetSubjectContextByUserID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Same(t, subject, got)
}
