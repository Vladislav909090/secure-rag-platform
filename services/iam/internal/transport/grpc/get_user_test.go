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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserServiceGetUserDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		GetUser(mock.Anything, "u1").
		RunAndReturn(func(ctx context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "u1", userID)

			return &model.UserView{ID: "u1", Login: "alice", IsActive: true, Roles: []string{usecase.RoleUser}}, nil
		})

	resp, err := (&UserServiceServerImpl{svc: uc}).GetUser(authContext("token"), &pb.GetUserRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUser().GetId())
	assert.Equal(t, "alice", resp.GetUser().GetLogin())
}

func TestUserServiceGetUserRejectsOtherUserForNonAdmin(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, (*model.SubjectContext)(nil), nil)

	_, err := (&UserServiceServerImpl{svc: uc}).GetUser(authContext("token"), &pb.GetUserRequest{UserId: "u2"})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
