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

func TestUserServiceUpdateUserMapsOptionalFields(t *testing.T) {
	t.Parallel()

	login := "alice2"
	password := "new-secret"
	active := false
	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}, nil, nil
	}
	mock.updateUser = func(ctx context.Context, input usecase.UpdateUserInput) (*model.UserView, error) {
		assert.Equal(t, "u1", input.UserID)
		require.NotNil(t, input.Login)
		assert.Equal(t, login, *input.Login)
		require.NotNil(t, input.Password)
		assert.Equal(t, password, *input.Password)
		require.NotNil(t, input.IsActive)
		assert.Equal(t, active, *input.IsActive)

		return &model.UserView{ID: input.UserID, Login: *input.Login, IsActive: *input.IsActive}, nil
	}

	resp, err := (&UserServiceServerImpl{svc: mock}).UpdateUser(authContext("token"), &pb.UpdateUserRequest{
		UserId:   "u1",
		Login:    &login,
		Password: &password,
		IsActive: &active,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUser().GetId())
	assert.Equal(t, login, resp.GetUser().GetLogin())
}
