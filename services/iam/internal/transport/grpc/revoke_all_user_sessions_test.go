package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionServiceRevokeAllUserSessionsMapsResult(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}
	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return principal, nil, nil
	}
	mock.revokeAllUserSessions = func(ctx context.Context, got *usecase.Principal, userID string) (*usecase.LogoutAllResult, error) {
		assert.Same(t, principal, got)
		assert.Equal(t, "u1", userID)

		return &usecase.LogoutAllResult{UserID: userID, RevokedCount: 2, CtxVer: 13}, nil
	}

	resp, err := (&SessionServiceServerImpl{svc: mock}).RevokeAllUserSessions(authContext("token"), &pb.RevokeAllUserSessionsRequest{UserId: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(2), resp.GetRevokedCount())
	assert.Equal(t, int64(13), resp.GetCtxVer())
}
