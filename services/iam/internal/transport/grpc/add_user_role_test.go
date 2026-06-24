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

func TestRoleServiceAddUserRoleUsesAdminPrincipal(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		AddUserRole(mock.Anything, "u1", usecase.RoleKnowledgeEditor, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string, roleCode string, assignedBy *string) ([]*model.Role, int64, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, usecase.RoleKnowledgeEditor, roleCode)
			require.NotNil(t, assignedBy)
			assert.Equal(t, "admin", *assignedBy)

			return []*model.Role{{ID: 2, Code: roleCode}}, 6, nil
		})

	resp, err := (&RoleServiceServerImpl{svc: uc}).AddUserRole(authContext("token"), &pb.AddUserRoleRequest{
		UserId:   "u1",
		RoleCode: usecase.RoleKnowledgeEditor,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(6), resp.GetCtxVer())
	require.Len(t, resp.GetRoles(), 1)
	assert.Equal(t, usecase.RoleKnowledgeEditor, resp.GetRoles()[0].GetCode())
}
