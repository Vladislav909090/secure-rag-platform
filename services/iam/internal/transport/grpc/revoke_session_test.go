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

func TestSessionServiceRevokeSessionUsesUsecase(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}
	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return principal, nil, nil
	}
	mock.revokeSession = func(ctx context.Context, got *usecase.Principal, sessionID string) (bool, error) {
		assert.Same(t, principal, got)
		assert.Equal(t, "s1", sessionID)

		return true, nil
	}

	resp, err := (&SessionServiceServerImpl{svc: mock}).RevokeSession(authContext("token"), &pb.RevokeSessionRequest{SessionId: "s1"})
	require.NoError(t, err)
	assert.True(t, resp.GetRevoked())
}
