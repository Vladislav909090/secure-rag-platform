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

func TestAuthServiceLogoutAllMapsResult(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(principal, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		LogoutAll(mock.Anything, principal, "u1").
		RunAndReturn(func(ctx context.Context, got *usecase.Principal, userID string) (*usecase.LogoutAllResult, error) {
			assert.Same(t, principal, got)
			assert.Equal(t, "u1", userID)

			return &usecase.LogoutAllResult{UserID: userID, RevokedCount: 3, CtxVer: 9}, nil
		})

	resp, err := (&AuthServiceServerImpl{svc: uc}).LogoutAll(authContext("token"), &pb.LogoutAllRequest{UserId: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(3), resp.GetRevokedCount())
	assert.Equal(t, int64(9), resp.GetCtxVer())
}
