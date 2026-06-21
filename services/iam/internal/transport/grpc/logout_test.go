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

func TestAuthServiceLogoutAuthenticatesAndRevokesCurrentSession(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "u1", SessionID: "s1", Roles: []string{usecase.RoleUser}}
	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		assert.Equal(t, "access-token", accessToken)

		return principal, nil, nil
	}
	mock.logout = func(ctx context.Context, got *usecase.Principal, sessionID string) (bool, error) {
		assert.Same(t, principal, got)
		assert.Empty(t, sessionID)

		return true, nil
	}

	resp, err := (&AuthServiceServerImpl{svc: mock}).Logout(authContext("access-token"), &pb.LogoutRequest{})
	require.NoError(t, err)
	assert.True(t, resp.GetRevoked())
}
