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

func TestAuthServiceLogoutAuthenticatesAndRevokesCurrentSession(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "u1", SessionID: "s1", Roles: []string{usecase.RoleUser}}
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "access-token").
		Return(principal, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		Logout(mock.Anything, principal, "").
		RunAndReturn(func(ctx context.Context, got *usecase.Principal, sessionID string) (bool, error) {
			assert.Same(t, principal, got)
			assert.Empty(t, sessionID)

			return true, nil
		})

	resp, err := (&AuthServiceServerImpl{svc: uc}).Logout(authContext("access-token"), &pb.LogoutRequest{})
	require.NoError(t, err)
	assert.True(t, resp.GetRevoked())
}
