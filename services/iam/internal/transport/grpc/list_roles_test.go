package grpc

import (
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoleServiceListRolesRequiresAuthentication(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		ListRoles(mock.Anything).
		Return([]*model.Role{{ID: 1, Code: usecase.RoleUser, Name: "User"}}, nil)

	resp, err := (&RoleServiceServerImpl{svc: uc}).ListRoles(authContext("token"), &pb.ListRolesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetRoles(), 1)
	assert.Equal(t, usecase.RoleUser, resp.GetRoles()[0].GetCode())
}
