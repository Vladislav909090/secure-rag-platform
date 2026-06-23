package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIAMRepo struct {
	t *testing.T

	addUserRole              func(context.Context, string, string, *string) ([]*model.Role, error)
	createSession            func(context.Context, repository.CreateSessionInput) (*model.UserSession, error)
	createUser               func(context.Context, repository.CreateUserInput) (*model.User, error)
	deleteUserAttributeKey   func(context.Context, string, string, *string) (map[string]any, error)
	getSessionByID           func(context.Context, string) (*model.UserSession, error)
	getSessionByRefreshHash  func(context.Context, string) (*model.UserSession, error)
	getSubjectContext        func(context.Context, string) (*model.SubjectContext, error)
	getUserAttributes        func(context.Context, string) (map[string]any, error)
	getUserByID              func(context.Context, string) (*model.User, error)
	getUserByLogin           func(context.Context, string) (*model.User, error)
	getUserRoles             func(context.Context, string) ([]*model.Role, error)
	getUserView              func(context.Context, string) (*model.UserView, error)
	hasUserWithRole          func(context.Context, string) (bool, error)
	incrementContextVersion  func(context.Context, string) (int64, error)
	listActiveUserSessions   func(context.Context, string) ([]*model.UserSession, error)
	listRoles                func(context.Context) ([]*model.Role, error)
	listUserViews            func(context.Context) ([]*model.UserView, error)
	removeUserRole           func(context.Context, string, string) ([]*model.Role, error)
	replaceUserAttributes    func(context.Context, string, map[string]any, *string) (map[string]any, error)
	revokeAllUserSessions    func(context.Context, string) (int64, error)
	revokeSession            func(context.Context, string) (bool, error)
	rotateSessionRefreshHash func(context.Context, string, string, time.Time) (*model.UserSession, error)
	setUserRoles             func(context.Context, string, []string, *string) ([]*model.Role, error)
	updateUser               func(context.Context, repository.UpdateUserInput) (*model.User, error)
}

var _ iamRepo = (*mockIAMRepo)(nil)

func (m *mockIAMRepo) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected repo call", "unexpected repo call: %s", name)
	}
	panic(fmt.Sprintf("unexpected repo call: %s", name))
}

func (m *mockIAMRepo) AddUserRole(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, error) {
	if m.addUserRole == nil {
		m.unexpected("AddUserRole")
	}

	return m.addUserRole(ctx, userID, roleCode, assignedBy)
}

func (m *mockIAMRepo) CreateSession(ctx context.Context, input repository.CreateSessionInput) (*model.UserSession, error) {
	if m.createSession == nil {
		m.unexpected("CreateSession")
	}

	return m.createSession(ctx, input)
}

func (m *mockIAMRepo) CreateUser(ctx context.Context, input repository.CreateUserInput) (*model.User, error) {
	if m.createUser == nil {
		m.unexpected("CreateUser")
	}

	return m.createUser(ctx, input)
}

func (m *mockIAMRepo) DeleteUserAttributeKey(ctx context.Context, userID string, key string, updatedBy *string) (map[string]any, error) {
	if m.deleteUserAttributeKey == nil {
		m.unexpected("DeleteUserAttributeKey")
	}

	return m.deleteUserAttributeKey(ctx, userID, key, updatedBy)
}

func (m *mockIAMRepo) GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	if m.getSessionByID == nil {
		m.unexpected("GetSessionByID")
	}

	return m.getSessionByID(ctx, sessionID)
}

func (m *mockIAMRepo) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (*model.UserSession, error) {
	if m.getSessionByRefreshHash == nil {
		m.unexpected("GetSessionByRefreshHash")
	}

	return m.getSessionByRefreshHash(ctx, refreshHash)
}

func (m *mockIAMRepo) GetSubjectContext(ctx context.Context, userID string) (*model.SubjectContext, error) {
	if m.getSubjectContext == nil {
		m.unexpected("GetSubjectContext")
	}

	return m.getSubjectContext(ctx, userID)
}

func (m *mockIAMRepo) GetUserAttributes(ctx context.Context, userID string) (map[string]any, error) {
	if m.getUserAttributes == nil {
		m.unexpected("GetUserAttributes")
	}

	return m.getUserAttributes(ctx, userID)
}

func (m *mockIAMRepo) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	if m.getUserByID == nil {
		m.unexpected("GetUserByID")
	}

	return m.getUserByID(ctx, userID)
}

func (m *mockIAMRepo) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if m.getUserByLogin == nil {
		m.unexpected("GetUserByLogin")
	}

	return m.getUserByLogin(ctx, login)
}

func (m *mockIAMRepo) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	if m.getUserRoles == nil {
		m.unexpected("GetUserRoles")
	}

	return m.getUserRoles(ctx, userID)
}

func (m *mockIAMRepo) GetUserView(ctx context.Context, userID string) (*model.UserView, error) {
	if m.getUserView == nil {
		m.unexpected("GetUserView")
	}

	return m.getUserView(ctx, userID)
}

func (m *mockIAMRepo) HasUserWithRole(ctx context.Context, roleCode string) (bool, error) {
	if m.hasUserWithRole == nil {
		m.unexpected("HasUserWithRole")
	}

	return m.hasUserWithRole(ctx, roleCode)
}

func (m *mockIAMRepo) IncrementContextVersion(ctx context.Context, userID string) (int64, error) {
	if m.incrementContextVersion == nil {
		m.unexpected("IncrementContextVersion")
	}

	return m.incrementContextVersion(ctx, userID)
}

func (m *mockIAMRepo) ListActiveUserSessions(ctx context.Context, userID string) ([]*model.UserSession, error) {
	if m.listActiveUserSessions == nil {
		m.unexpected("ListActiveUserSessions")
	}

	return m.listActiveUserSessions(ctx, userID)
}

func (m *mockIAMRepo) ListRoles(ctx context.Context) ([]*model.Role, error) {
	if m.listRoles == nil {
		m.unexpected("ListRoles")
	}

	return m.listRoles(ctx)
}

func (m *mockIAMRepo) ListUserViews(ctx context.Context) ([]*model.UserView, error) {
	if m.listUserViews == nil {
		m.unexpected("ListUserViews")
	}

	return m.listUserViews(ctx)
}

func (m *mockIAMRepo) RemoveUserRole(ctx context.Context, userID string, roleCode string) ([]*model.Role, error) {
	if m.removeUserRole == nil {
		m.unexpected("RemoveUserRole")
	}

	return m.removeUserRole(ctx, userID, roleCode)
}

func (m *mockIAMRepo) ReplaceUserAttributes(ctx context.Context, userID string, attrs map[string]any, updatedBy *string) (map[string]any, error) {
	if m.replaceUserAttributes == nil {
		m.unexpected("ReplaceUserAttributes")
	}

	return m.replaceUserAttributes(ctx, userID, attrs, updatedBy)
}

func (m *mockIAMRepo) RevokeAllUserSessions(ctx context.Context, userID string) (int64, error) {
	if m.revokeAllUserSessions == nil {
		m.unexpected("RevokeAllUserSessions")
	}

	return m.revokeAllUserSessions(ctx, userID)
}

func (m *mockIAMRepo) RevokeSession(ctx context.Context, sessionID string) (bool, error) {
	if m.revokeSession == nil {
		m.unexpected("RevokeSession")
	}

	return m.revokeSession(ctx, sessionID)
}

func (m *mockIAMRepo) RotateSessionRefreshHash(ctx context.Context, sessionID string, newRefreshHash string, expiresAt time.Time) (*model.UserSession, error) {
	if m.rotateSessionRefreshHash == nil {
		m.unexpected("RotateSessionRefreshHash")
	}

	return m.rotateSessionRefreshHash(ctx, sessionID, newRefreshHash, expiresAt)
}

func (m *mockIAMRepo) SetUserRoles(ctx context.Context, userID string, roleCodes []string, assignedBy *string) ([]*model.Role, error) {
	if m.setUserRoles == nil {
		m.unexpected("SetUserRoles")
	}

	return m.setUserRoles(ctx, userID, roleCodes, assignedBy)
}

func (m *mockIAMRepo) UpdateUser(ctx context.Context, input repository.UpdateUserInput) (*model.User, error) {
	if m.updateUser == nil {
		m.unexpected("UpdateUser")
	}

	return m.updateUser(ctx, input)
}

func newIAMTestUsecase(repo iamRepo) *IAMUsecase {
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

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})
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
	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(_ context.Context, userID string) (*model.SubjectContext, error) {
			assert.Equal(t, "u1", userID)

			return subject, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.GetSubjectContext(context.Background(), "u1")
	require.NoError(t, err)
	assert.Same(t, subject, got)
}

func TestIAMUsecaseGetSubjectContextRejectsEmptyUserID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.GetSubjectContext(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseAuthenticateAccessTokenRejectsContextVersionMismatch(t *testing.T) {
	t.Parallel()

	tokenSubject := iamTestSubject("u1")
	repoSubject := iamTestSubject("u1")
	repoSubject.CtxVer = 4
	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(context.Context, string) (*model.SubjectContext, error) {
			return repoSubject, nil
		},
	}
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
	repo := &mockIAMRepo{
		t: t,
		getSubjectContext: func(context.Context, string) (*model.SubjectContext, error) {
			return subject, nil
		},
		getSessionByID: func(context.Context, string) (*model.UserSession, error) {
			return session, nil
		},
	}
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

	repo := &mockIAMRepo{
		t: t,
		hasUserWithRole: func(_ context.Context, roleCode string) (bool, error) {
			assert.Equal(t, RoleSuperAdmin, roleCode)

			return true, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), "admin", "password")
	require.NoError(t, err)
	assert.Empty(t, password)
	assert.False(t, created)
}

func TestIAMUsecaseBootstrapSuperAdminCreatesGeneratedUser(t *testing.T) {
	t.Parallel()

	view := iamTestUserView("admin-id")
	repo := &mockIAMRepo{
		t: t,
		hasUserWithRole: func(_ context.Context, roleCode string) (bool, error) {
			assert.Equal(t, RoleSuperAdmin, roleCode)

			return false, nil
		},
		createUser: func(_ context.Context, input repository.CreateUserInput) (*model.User, error) {
			assert.Equal(t, "admin@example.com", input.Login)
			assert.NotEmpty(t, input.PasswordHash)
			assert.True(t, input.IsActive)
			assert.Equal(t, []string{RoleSuperAdmin, RoleUser}, input.RoleCodes)
			assert.Equal(t, map[string]any{}, input.Attributes)

			return iamTestUser("admin-id"), nil
		},
		getUserView: func(_ context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "admin-id", userID)

			return view, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), " admin@example.com ", "")
	require.NoError(t, err)
	assert.NotEmpty(t, password)
	assert.True(t, created)
}

func TestIAMUsecaseBootstrapSuperAdminRejectsEmptyLogin(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		hasUserWithRole: func(context.Context, string) (bool, error) {
			return false, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	password, created, err := uc.BootstrapSuperAdmin(context.Background(), " ", "password")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Empty(t, password)
	assert.False(t, created)
}
