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

func TestRoleServiceGetUserRolesDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		GetUserRoles(mock.Anything, "u1").
		RunAndReturn(func(ctx context.Context, userID string) ([]*model.Role, int64, error) {
			assert.Equal(t, "u1", userID)

			return []*model.Role{{ID: 1, Code: usecase.RoleUser}}, 5, nil
		})

	resp, err := (&RoleServiceServerImpl{svc: uc}).GetUserRoles(authContext("token"), &pb.GetUserRolesRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(5), resp.GetCtxVer())
	require.Len(t, resp.GetRoles(), 1)
}
