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

func TestRoleServiceListRolesRequiresAuthentication(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
	}
	mock.listRoles = func(ctx context.Context) ([]*model.Role, error) {
		return []*model.Role{{ID: 1, Code: usecase.RoleUser, Name: "User"}}, nil
	}

	resp, err := (&RoleServiceServerImpl{svc: mock}).ListRoles(authContext("token"), &pb.ListRolesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetRoles(), 1)
	assert.Equal(t, usecase.RoleUser, resp.GetRoles()[0].GetCode())
}
