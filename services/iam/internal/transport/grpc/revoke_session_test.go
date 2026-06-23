package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSessionServiceRevokeSessionUsesUsecase(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(principal, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		RevokeSession(mock.Anything, principal, "s1").
		RunAndReturn(func(ctx context.Context, got *usecase.Principal, sessionID string) (bool, error) {
			assert.Same(t, principal, got)
			assert.Equal(t, "s1", sessionID)

			return true, nil
		})

	resp, err := (&SessionServiceServerImpl{svc: uc}).RevokeSession(authContext("token"), &pb.RevokeSessionRequest{SessionId: "s1"})
	require.NoError(t, err)
	assert.True(t, resp.GetRevoked())
}
