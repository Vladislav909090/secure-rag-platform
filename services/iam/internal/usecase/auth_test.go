package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseLoginIssuesTokenPair(t *testing.T) {
	t.Parallel()

	user := iamTestUser("u1")
	subject := iamTestSubject("u1")
	repo := &mockIAMRepo{
		t: t,
		getUserByLogin: func(_ context.Context, login string) (*model.User, error) {
			assert.Equal(t, "user@example.com", login)

			return user, nil
		},
		getSubjectContext: func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return subject, nil
		},
		createSession: func(_ context.Context, input repository.CreateSessionInput) (*model.UserSession, error) {
			assert.Equal(t, "u1", input.UserID)
			assert.NotEmpty(t, input.RefreshTokenHash)
			assert.True(t, input.ExpiresAt.After(time.Now().UTC()))

			return iamTestSession("s1", "u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.Login(context.Background(), LoginInput{Login: " user@example.com ", Password: "password"})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.Equal(t, tokenTypeBearer, got.TokenType)
	assert.Positive(t, got.ExpiresIn)
}

func TestIAMUsecaseLoginRejectsBadPassword(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByLogin: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.Login(context.Background(), LoginInput{Login: "user@example.com", Password: "bad"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, got)
}

func TestIAMUsecaseLoginRejectsInactiveUser(t *testing.T) {
	t.Parallel()

	user := iamTestUser("u1")
	user.IsActive = false
	repo := &mockIAMRepo{
		t: t,
		getUserByLogin: func(context.Context, string) (*model.User, error) {
			return user, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.Login(context.Background(), LoginInput{Login: "user@example.com", Password: "password"})
	require.ErrorIs(t, err, ErrInactiveUser)
	assert.Nil(t, got)
}

func TestIAMUsecaseLoginRejectsEmptyInput(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.Login(context.Background(), LoginInput{Login: " ", Password: "password"})
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseRefreshTokenRotatesSession(t *testing.T) {
	t.Parallel()

	refreshRaw := "refresh-token"
	refreshHash := hashOpaqueToken(refreshRaw)
	repo := &mockIAMRepo{
		t: t,
		getSessionByRefreshHash: func(_ context.Context, gotHash string) (*model.UserSession, error) {
			assert.Equal(t, refreshHash, gotHash)

			return iamTestSession("s1", "u1"), nil
		},
		getUserByID: func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u1", userID)

			return iamTestUser("u1"), nil
		},
		getSubjectContext: func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return iamTestSubject("u1"), nil
		},
		rotateSessionRefreshHash: func(_ context.Context, sessionID string, newRefreshHash string, expiresAt time.Time) (*model.UserSession, error) {
			assert.Equal(t, "s1", sessionID)
			assert.NotEmpty(t, newRefreshHash)
			assert.NotEqual(t, refreshHash, newRefreshHash)
			assert.True(t, expiresAt.After(time.Now().UTC()))

			return iamTestSession("s1", "u1"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: refreshRaw})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.Equal(t, tokenTypeBearer, got.TokenType)
}

func TestIAMUsecaseRefreshTokenRejectsRevokedSession(t *testing.T) {
	t.Parallel()

	revokedAt := time.Now().UTC()
	session := iamTestSession("s1", "u1")
	session.RevokedAt = &revokedAt
	repo := &mockIAMRepo{
		t: t,
		getSessionByRefreshHash: func(context.Context, string) (*model.UserSession, error) {
			return session, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh-token"})
	require.ErrorIs(t, err, ErrSessionRevoked)
	assert.Nil(t, got)
}

func TestIAMUsecaseRefreshTokenRejectsExpiredSession(t *testing.T) {
	t.Parallel()

	session := iamTestSession("s1", "u1")
	session.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	repo := &mockIAMRepo{
		t: t,
		getSessionByRefreshHash: func(context.Context, string) (*model.UserSession, error) {
			return session, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh-token"})
	require.ErrorIs(t, err, ErrSessionExpired)
	assert.Nil(t, got)
}

func TestIAMUsecaseRefreshTokenRejectsMissingUser(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSessionByRefreshHash: func(context.Context, string) (*model.UserSession, error) {
			return iamTestSession("s1", "u1"), nil
		},
		getUserByID: func(context.Context, string) (*model.User, error) {
			return nil, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh-token"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, got)
}

func TestIAMUsecaseRefreshTokenMapsRotateNotFoundToRevoked(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSessionByRefreshHash: func(context.Context, string) (*model.UserSession, error) {
			return iamTestSession("s1", "u1"), nil
		},
		getUserByID: func(context.Context, string) (*model.User, error) {
			return iamTestUser("u1"), nil
		},
		getSubjectContext: func(context.Context, string) (*model.SubjectContext, error) {
			return iamTestSubject("u1"), nil
		},
		rotateSessionRefreshHash: func(context.Context, string, string, time.Time) (*model.UserSession, error) {
			return nil, repository.ErrNotFound
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh-token"})
	require.ErrorIs(t, err, ErrSessionRevoked)
	assert.Nil(t, got)
}

func TestIAMUsecaseLogoutAllowsAdminToRevokeOtherUserSession(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSessionByID: func(_ context.Context, sessionID string) (*model.UserSession, error) {
			assert.Equal(t, "s2", sessionID)

			return iamTestSession("s2", "u2"), nil
		},
		revokeSession: func(_ context.Context, sessionID string) (bool, error) {
			assert.Equal(t, "s2", sessionID)

			return true, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	revoked, err := uc.Logout(context.Background(), &Principal{UserID: "u1", SessionID: "s1", Roles: []string{RoleAccessAdmin}}, "s2")
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestIAMUsecaseLogoutRejectsOtherUserForNonAdmin(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSessionByID: func(context.Context, string) (*model.UserSession, error) {
			return iamTestSession("s2", "u2"), nil
		},
	}
	uc := newIAMTestUsecase(repo)

	revoked, err := uc.Logout(context.Background(), &Principal{UserID: "u1", SessionID: "s1", Roles: []string{RoleUser}}, "s2")
	require.ErrorIs(t, err, ErrForbidden)
	assert.False(t, revoked)
}

func TestIAMUsecaseLogoutRejectsNilPrincipal(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	revoked, err := uc.Logout(context.Background(), nil, "s1")
	require.ErrorIs(t, err, ErrUnauthorized)
	assert.False(t, revoked)
}

func TestIAMUsecaseLogoutReturnsNotFoundForMissingSession(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getSessionByID: func(context.Context, string) (*model.UserSession, error) {
			return nil, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	revoked, err := uc.Logout(context.Background(), &Principal{UserID: "u1", SessionID: "s1"}, "")
	require.ErrorIs(t, err, ErrNotFound)
	assert.False(t, revoked)
}

func TestIAMUsecaseLogoutAllRevokesTargetUserSessions(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(_ context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "u2", userID)

			return iamTestUser("u2"), nil
		},
		revokeAllUserSessions: func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u2", userID)

			return 3, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.LogoutAll(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleSuperAdmin}}, "u2")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u2", got.UserID)
	assert.Equal(t, int64(3), got.RevokedCount)
	assert.Equal(t, int64(3), got.CtxVer)
}

func TestIAMUsecaseLogoutAllRejectsOtherUserForNonAdmin(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.LogoutAll(context.Background(), &Principal{UserID: "u1", Roles: []string{RoleUser}}, "u2")
	require.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, got)
}

func TestIAMUsecaseLogoutAllReturnsNotFoundForMissingUser(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserByID: func(context.Context, string) (*model.User, error) {
			return nil, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.LogoutAll(context.Background(), &Principal{UserID: "u1"}, "")
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, got)
}

func TestIAMUsecaseGetMeReturnsPrincipalContext(t *testing.T) {
	t.Parallel()

	subject := iamTestSubject("u1")
	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(context.Context, string) (*model.SubjectContext, error) {
			return subject, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.GetMe(context.Background(), &Principal{UserID: "u1"})
	require.NoError(t, err)
	assert.Same(t, subject, got)
}

func TestIAMUsecaseGetMeRejectsNilPrincipal(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.GetMe(context.Background(), nil)
	require.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, got)
}
