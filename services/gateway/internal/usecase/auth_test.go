package usecase

import (
	"context"
	"testing"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAccessRoleHelpers(t *testing.T) {
	t.Parallel()

	require.NoError(t, requireAdmin(nil))
	require.NoError(t, requireAdmin(&iamv1.SubjectContext{Roles: []string{roleAccessAdmin}}))
	require.ErrorIs(t, requireAdmin(&iamv1.SubjectContext{Roles: []string{roleUser}}), ErrForbidden)

	require.NoError(t, requireDocumentEditor(nil))
	require.NoError(t, requireDocumentEditor(&iamv1.SubjectContext{Roles: []string{roleKnowledgeEditor}}))
	require.ErrorIs(t, requireDocumentEditor(&iamv1.SubjectContext{Roles: []string{roleAccessAdmin}}), ErrForbidden)
}

func TestOutgoingAuthContext(t *testing.T) {
	t.Parallel()

	ctx, err := outgoingAuthContext(context.Background(), " token ")
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	assert.Equal(t, []string{"Bearer token"}, md.Get("authorization"))

	_, err = outgoingAuthContext(context.Background(), " ")
	require.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuthProtoConversions(t *testing.T) {
	t.Parallel()

	tokens := tokenPairFromProto(&iamv1.TokenPairResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresIn:    30,
		TokenType:    "Bearer",
	})
	assert.Equal(t, "access", tokens.AccessToken)
	assert.Equal(t, "refresh", tokens.RefreshToken)
	assert.Equal(t, int64(30), tokens.ExpiresIn)
	assert.Equal(t, "Bearer", tokens.TokenType)
	assert.Empty(t, tokenPairFromProto(nil).AccessToken)

	attrs, err := structpb.NewStruct(map[string]any{"department": "legal"})
	require.NoError(t, err)
	subject := subjectFromProto(&iamv1.SubjectContext{
		UserId:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      []string{roleUser},
		Attributes: attrs,
		CtxVer:     7,
	})
	assert.Equal(t, "u1", subject.UserID)
	assert.Equal(t, "alice", subject.Login)
	assert.Equal(t, "legal", subject.Attributes["department"])
	assert.Equal(t, int64(7), subject.CtxVer)
	assert.Nil(t, subjectFromProto(nil))
}

func TestGatewayServiceLoginRefreshLogoutAndGetMe(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	deps.auth.EXPECT().
		Login(mock.Anything, mock.MatchedBy(func(req *iamv1.LoginRequest) bool {
			return req.GetLogin() == "alice" && req.GetPassword() == "secret"
		})).
		Return(&iamv1.TokenPairResponse{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 60, TokenType: "Bearer"}, nil)
	deps.auth.EXPECT().
		RefreshToken(mock.Anything, mock.MatchedBy(func(req *iamv1.RefreshTokenRequest) bool {
			return req.GetRefreshToken() == "old-refresh"
		})).
		Return(&iamv1.TokenPairResponse{AccessToken: "new-access", RefreshToken: "new-refresh", ExpiresIn: 30, TokenType: "Bearer"}, nil)
	deps.auth.EXPECT().
		Logout(mock.MatchedBy(hasBearer("access")), mock.Anything).
		Return(&iamv1.LogoutResponse{Revoked: true}, nil)
	deps.auth.EXPECT().
		GetMe(mock.MatchedBy(hasBearer("access")), mock.Anything).
		Return(&iamv1.GetMeResponse{Me: gatewaySubject(roleUser)}, nil)

	tokens, err := svc.Login(context.Background(), LoginRequest{Login: " alice ", Password: "secret"})
	require.NoError(t, err)
	assert.Equal(t, "access", tokens.AccessToken)

	tokens, err = svc.RefreshToken(context.Background(), " old-refresh ")
	require.NoError(t, err)
	assert.Equal(t, "new-access", tokens.AccessToken)

	revoked, err := svc.Logout(context.Background(), "access")
	require.NoError(t, err)
	assert.True(t, revoked)

	subject, err := svc.GetMe(context.Background(), "access")
	require.NoError(t, err)
	assert.Equal(t, "u1", subject.UserID)
}

func TestGatewayServiceAuthRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	svc, _ := newGatewayTestService(t)

	tokens, err := svc.Login(context.Background(), LoginRequest{Login: " ", Password: "secret"})
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, tokens)

	tokens, err = svc.RefreshToken(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, tokens)

	revoked, err := svc.Logout(context.Background(), " ")
	require.ErrorIs(t, err, ErrUnauthorized)
	assert.False(t, revoked)
}

func TestGatewayServiceValidateAccessTokenMapsInvalidResponses(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	deps.iam.EXPECT().
		ValidateAccessToken(mock.Anything, mock.Anything).
		Return(&iamv1.ValidateAccessTokenResponse{Valid: false}, nil)

	subject, err := svc.validateAccessToken(context.Background(), "token")
	require.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, subject)
}

func hasBearer(token string) func(context.Context) bool {
	return func(ctx context.Context) bool {
		md, ok := metadata.FromOutgoingContext(ctx)
		return ok && len(md.Get("authorization")) == 1 && md.Get("authorization")[0] == "Bearer "+token
	}
}

func TestGatewayServiceLoginMapsUpstreamError(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	deps.auth.EXPECT().
		Login(mock.Anything, mock.Anything).
		Return((*iamv1.TokenPairResponse)(nil), assert.AnError)

	tokens, err := svc.Login(context.Background(), LoginRequest{Login: "alice", Password: "secret"})
	require.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, tokens)
}
