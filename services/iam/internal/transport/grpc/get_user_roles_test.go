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

func TestRoleServiceGetUserRolesDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
	}
	mock.getUserRoles = func(ctx context.Context, userID string) ([]*model.Role, int64, error) {
		assert.Equal(t, "u1", userID)

		return []*model.Role{{ID: 1, Code: usecase.RoleUser}}, 5, nil
	}

	resp, err := (&RoleServiceServerImpl{svc: mock}).GetUserRoles(authContext("token"), &pb.GetUserRolesRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(5), resp.GetCtxVer())
	require.Len(t, resp.GetRoles(), 1)
}
