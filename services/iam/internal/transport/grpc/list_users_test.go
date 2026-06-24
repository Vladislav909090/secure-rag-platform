package grpc

import (
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserServiceListUsersRequiresAdminAndMapsUsers(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "admin-token").
		Return(&usecase.Principal{
			UserID: "admin",
			Roles:  []string{usecase.RoleAccessAdmin},
		}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		ListUsers(mock.Anything).
		Return([]*model.UserView{
			{ID: "u1", Login: "alice", IsActive: true, Roles: []string{usecase.RoleUser}},
			{ID: "u2", Login: "bob", IsActive: false, Roles: []string{usecase.RoleUser}},
		}, nil)

	resp, err := (&UserServiceServerImpl{svc: uc}).ListUsers(authContext("admin-token"), &pb.ListUsersRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetUsers(), 2)
	assert.Equal(t, "alice", resp.GetUsers()[0].GetLogin())
	assert.Equal(t, "bob", resp.GetUsers()[1].GetLogin())
}

func TestUserServiceListUsersRejectsNonAdmin(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "user-token").
		Return(&usecase.Principal{
			UserID: "u1",
			Roles:  []string{usecase.RoleUser},
		}, (*model.SubjectContext)(nil), nil)

	_, err := (&UserServiceServerImpl{svc: uc}).ListUsers(authContext("user-token"), &pb.ListUsersRequest{})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
