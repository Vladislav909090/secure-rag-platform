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

func TestRoleServiceSetUserRolesUsesAdminPrincipal(t *testing.T) {
	t.Parallel()

	roleCodes := []string{usecase.RoleUser, usecase.RoleAccessAdmin}
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleSuperAdmin}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		SetUserRoles(mock.Anything, "u1", roleCodes, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string, gotRoleCodes []string, assignedBy *string) ([]*model.Role, int64, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, roleCodes, gotRoleCodes)
			require.NotNil(t, assignedBy)
			assert.Equal(t, "admin", *assignedBy)

			return []*model.Role{{ID: 1, Code: usecase.RoleUser}, {ID: 3, Code: usecase.RoleAccessAdmin}}, 8, nil
		})

	resp, err := (&RoleServiceServerImpl{svc: uc}).SetUserRoles(authContext("token"), &pb.SetUserRolesRequest{
		UserId:    "u1",
		RoleCodes: roleCodes,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(8), resp.GetCtxVer())
	require.Len(t, resp.GetRoles(), 2)
}
