package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newIAMTestUsecase(repo IAMRepo) *IAMUsecase {
	cfg := DefaultConfig()
	cfg.JWTSecret = "test-secret"

	return &IAMUsecase{repo: repo, cfg: cfg}
}

func iamTestUser(id string) *model.User {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	hash, err := hashPassword("password")
	if err != nil {
		panic(err)
	}

	return &model.User{
		ID:           id,
		Login:        "user@example.com",
		PasswordHash: hash,
		IsActive:     true,
		CtxVer:       3,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func iamTestSubject(userID string) *model.SubjectContext {
	return &model.SubjectContext{
		UserID:     userID,
		Login:      "user@example.com",
		IsActive:   true,
		Roles:      []string{RoleUser},
		Attributes: map[string]any{"department": "search"},
		CtxVer:     3,
	}
}

func iamTestUserView(userID string) *model.UserView {
	user := iamTestUser(userID)

	return &model.UserView{
		ID:         user.ID,
		Login:      user.Login,
		IsActive:   user.IsActive,
		CtxVer:     user.CtxVer,
		Roles:      []string{RoleUser},
		Attributes: map[string]any{"department": "search"},
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func iamTestRole(code string) *model.Role {
	return &model.Role{ID: 1, Code: code, Name: code}
}

func iamTestSession(sessionID string, userID string) *model.UserSession {
	now := time.Now().UTC()

	return &model.UserSession{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: "refresh-hash",
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestUsecaseNormalizeRoleCodes(t *testing.T) {
	t.Parallel()

	got, err := normalizeRoleCodes([]string{" knowledge_editor ", "user", "user", ""})
	require.NoError(t, err)
	want := []string{RoleKnowledgeEditor, RoleUser}
	assert.Equal(t, want, got)

	got, err = normalizeRoleCodes(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{RoleUser}, got)

	_, err = normalizeRoleCodes([]string{"bad"})
	require.ErrorIs(t, err, ErrInvalidArgument)
}

func TestIAMPasswordAndTokenHelpers(t *testing.T) {
	t.Parallel()

	hash, err := hashPassword("secret")
	require.NoError(t, err)
	assert.True(t, checkPassword(hash, "secret"))
	assert.False(t, checkPassword(hash, "other"))
	_, err = hashPassword(" ")
	require.ErrorIs(t, err, ErrInvalidArgument)

	firstTokenHash := hashOpaqueToken("token")
	secondTokenHash := hashOpaqueToken("token")
	assert.Equal(t, firstTokenHash, secondTokenHash)

	token, err := randomToken(8)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRoleHelpers(t *testing.T) {
	t.Parallel()

	roles := []string{RoleUser, RoleAccessAdmin}
	assert.True(t, hasRole(roles, RoleUser))
	assert.True(t, hasAnyRole(roles, RoleSuperAdmin, RoleAccessAdmin))
	assert.False(t, hasAnyRole(roles, RoleSuperAdmin, RoleKnowledgeEditor))
}

func TestMergeAttributes(t *testing.T) {
	t.Parallel()

	got := mergeAttributes(map[string]any{"a": 1, "b": 2}, map[string]any{"b": 3, "c": 4})
	assert.Equal(t, map[string]any{"a": 1, "b": 3, "c": 4}, got)
}

func TestSubjectCacheKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "iam:subjectctx:u1", subjectCacheKey("u1"))
}

func TestNewIAMUsecaseAppliesDefaults(t *testing.T) {
	t.Parallel()

	uc := NewIAMUsecase(nil, nil, Config{}, nil)
	require.NotNil(t, uc)
	assert.NotEmpty(t, uc.cfg.JWTSecret)
	assert.Equal(t, DefaultConfig().JWTIssuer, uc.cfg.JWTIssuer)
	assert.Equal(t, DefaultConfig().JWTAudience, uc.cfg.JWTAudience)
	assert.Equal(t, DefaultConfig().AccessTokenTTL, uc.cfg.AccessTokenTTL)
	assert.Equal(t, DefaultConfig().RefreshTokenTTL, uc.cfg.RefreshTokenTTL)
	assert.Equal(t, DefaultConfig().SubjectCacheTTL, uc.cfg.SubjectCacheTTL)
	assert.Equal(t, DefaultConfig().LoginRateLimit, uc.cfg.LoginRateLimit)
	assert.Equal(t, DefaultConfig().RefreshRateLimit, uc.cfg.RefreshRateLimit)
}

func TestIAMUsecaseIssueAndParseAccessToken(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))
	subject := iamTestSubject("u1")

	token, expiresIn, tokenID, err := uc.issueAccessToken(subject, "s1")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Positive(t, expiresIn)
	assert.NotEmpty(t, tokenID)

	claims, err := uc.parseAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "u1", claims.Subject)
	assert.Equal(t, "s1", claims.SessionID)
	assert.Equal(t, subject.Roles, claims.Roles)
	assert.Equal(t, subject.CtxVer, claims.CtxVer)
	assert.Equal(t, tokenID, claims.ID)
}

func TestIAMUsecaseGetSubjectContextReturnsRepoContext(t *testing.T) {
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

	got, err := uc.GetSubjectContext(context.Background(), "u1")
	require.NoError(t, err)
	assert.Same(t, subject, got)
}

func TestIAMUsecaseGetSubjectContextRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(NewMockIAMRepo(t))

	got, err := uc.GetSubjectContext(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseAuthenticateAccessTokenRejectsContextVersionMismatch(t *testing.T) {
	t.Parallel()

	tokenSubject := iamTestSubject("u1")
	repoSubject := iamTestSubject("u1")
	repoSubject.CtxVer = 4
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSubjectContext(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.SubjectContext, error) {
			return repoSubject, nil
		})
	uc := newIAMTestUsecase(repo)
	token, _, _, err := uc.issueAccessToken(tokenSubject, "s1")
	require.NoError(t, err)

	principal, subject, err := uc.AuthenticateAccessToken(context.Background(), token)
	require.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, principal)
	assert.Nil(t, subject)
}

func TestIAMUsecaseAuthenticateAccessTokenRejectsRevokedSession(t *testing.T) {
	t.Parallel()

	subject := iamTestSubject("u1")
	revokedAt := time.Now().UTC()
	session := iamTestSession("s1", "u1")
	session.RevokedAt = &revokedAt
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		GetSubjectContext(mock.Anything, "u1").
		RunAndReturn(func(context.Context, string) (*model.SubjectContext, error) {
			return subject, nil
		})
	repo.EXPECT().
		GetSessionByID(mock.Anything, "s1").
		RunAndReturn(func(context.Context, string) (*model.UserSession, error) {
			return session, nil
		})
	uc := newIAMTestUsecase(repo)
	token, _, _, err := uc.issueAccessToken(subject, "s1")
	require.NoError(t, err)

	principal, gotSubject, err := uc.AuthenticateAccessToken(context.Background(), token)
	require.ErrorIs(t, err, ErrSessionRevoked)
	assert.Nil(t, principal)
	assert.Nil(t, gotSubject)
}

func TestIAMUsecaseBootstrapSuperAdminSkipsExisting(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		HasUserWithRole(mock.Anything, RoleSuperAdmin).
		RunAndReturn(func(_ context.Context, roleCode string) (bool, error) {
			assert.Equal(t, RoleSuperAdmin, roleCode)

			return true, nil
		})
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), "admin", "password")
	require.NoError(t, err)
	assert.Empty(t, password)
	assert.False(t, created)
}

func TestIAMUsecaseBootstrapSuperAdminCreatesGeneratedUser(t *testing.T) {
	t.Parallel()

	view := iamTestUserView("admin-id")
	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		HasUserWithRole(mock.Anything, RoleSuperAdmin).
		RunAndReturn(func(_ context.Context, roleCode string) (bool, error) {
			assert.Equal(t, RoleSuperAdmin, roleCode)

			return false, nil
		})
	repo.EXPECT().
		CreateUser(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, input repository.CreateUserInput) (*model.User, error) {
			assert.Equal(t, "admin@example.com", input.Login)
			assert.NotEmpty(t, input.PasswordHash)
			assert.True(t, input.IsActive)
			assert.Equal(t, []string{RoleSuperAdmin, RoleUser}, input.RoleCodes)
			assert.Equal(t, map[string]any{}, input.Attributes)

			return iamTestUser("admin-id"), nil
		})
	repo.EXPECT().
		GetUserView(mock.Anything, "admin-id").
		RunAndReturn(func(_ context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "admin-id", userID)

			return view, nil
		})
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), " admin@example.com ", "")
	require.NoError(t, err)
	assert.NotEmpty(t, password)
	assert.True(t, created)
}

func TestIAMUsecaseBootstrapSuperAdminRejectsEmptyLogin(t *testing.T) {
	t.Parallel()

	repo := NewMockIAMRepo(t)
	repo.EXPECT().
		HasUserWithRole(mock.Anything, RoleSuperAdmin).
		RunAndReturn(func(context.Context, string) (bool, error) {
			return false, nil
		})
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), " ", "password")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Empty(t, password)
	assert.False(t, created)
}
