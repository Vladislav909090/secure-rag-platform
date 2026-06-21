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

func TestUserServiceGetUserDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
	}
	mock.getUser = func(ctx context.Context, userID string) (*model.UserView, error) {
		assert.Equal(t, "u1", userID)

		return &model.UserView{ID: "u1", Login: "alice", IsActive: true, Roles: []string{usecase.RoleUser}}, nil
	}

	resp, err := (&UserServiceServerImpl{svc: mock}).GetUser(authContext("token"), &pb.GetUserRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUser().GetId())
	assert.Equal(t, "alice", resp.GetUser().GetLogin())
}

func TestUserServiceGetUserRejectsOtherUserForNonAdmin(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{
		t: t,
		authenticateAccessToken: func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
			return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
		},
		getUser: func(ctx context.Context, userID string) (*model.UserView, error) {
			require.FailNow(t, "GetUser usecase should not be called")

			return nil, nil
		},
	}

	_, err := (&UserServiceServerImpl{svc: mock}).GetUser(authContext("token"), &pb.GetUserRequest{UserId: "u2"})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
