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

func TestAuthServiceGetMeUsesAuthenticatedPrincipal(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(principal, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		GetMe(mock.Anything, principal).
		RunAndReturn(func(ctx context.Context, got *usecase.Principal) (*model.SubjectContext, error) {
			assert.Same(t, principal, got)

			return &model.SubjectContext{UserID: "u1", Login: "alice", IsActive: true, Roles: []string{usecase.RoleUser}}, nil
		})

	resp, err := (&AuthServiceServerImpl{svc: uc}).GetMe(authContext("token"), &pb.GetMeRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetMe().GetUserId())
	assert.Equal(t, "alice", resp.GetMe().GetLogin())
}
