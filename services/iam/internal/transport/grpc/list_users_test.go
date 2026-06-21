package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserServiceListUsersRequiresAdminAndMapsUsers(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		assert.Equal(t, "admin-token", accessToken)

		return &usecase.Principal{
			UserID: "admin",
			Roles:  []string{usecase.RoleAccessAdmin},
		}, nil, nil
	}
	mock.listUsers = func(ctx context.Context) ([]*model.UserView, error) {
		return []*model.UserView{
			{ID: "u1", Login: "alice", IsActive: true, Roles: []string{usecase.RoleUser}},
			{ID: "u2", Login: "bob", IsActive: false, Roles: []string{usecase.RoleUser}},
		}, nil
	}

	resp, err := (&UserServiceServerImpl{svc: mock}).ListUsers(authContext("admin-token"), &pb.ListUsersRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetUsers(), 2)
	assert.Equal(t, "alice", resp.GetUsers()[0].GetLogin())
	assert.Equal(t, "bob", resp.GetUsers()[1].GetLogin())
}

func TestUserServiceListUsersRejectsNonAdmin(t *testing.T) {
	t.Parallel()

	listCalled := false
	mock := &mockIAMUsecase{
		t: t,
		authenticateAccessToken: func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
			return &usecase.Principal{
				UserID: "u1",
				Roles:  []string{usecase.RoleUser},
			}, nil, nil
		},
		listUsers: func(ctx context.Context) ([]*model.UserView, error) {
			listCalled = true
			return nil, nil
		},
	}

	_, err := (&UserServiceServerImpl{svc: mock}).ListUsers(authContext("user-token"), &pb.ListUsersRequest{})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
	assert.False(t, listCalled)
}
