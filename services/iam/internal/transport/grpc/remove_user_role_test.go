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

func TestRoleServiceRemoveUserRoleUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}, nil, nil
	}
	mock.removeUserRole = func(ctx context.Context, userID string, roleCode string) ([]*model.Role, int64, error) {
		assert.Equal(t, "u1", userID)
		assert.Equal(t, usecase.RoleKnowledgeEditor, roleCode)

		return []*model.Role{{ID: 1, Code: usecase.RoleUser}}, 7, nil
	}

	resp, err := (&RoleServiceServerImpl{svc: mock}).RemoveUserRole(authContext("token"), &pb.RemoveUserRoleRequest{
		UserId:   "u1",
		RoleCode: usecase.RoleKnowledgeEditor,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(7), resp.GetCtxVer())
	require.Len(t, resp.GetRoles(), 1)
	assert.Equal(t, usecase.RoleUser, resp.GetRoles()[0].GetCode())
}
