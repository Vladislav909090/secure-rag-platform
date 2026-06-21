package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUserServiceCreateUserMapsRequestAndCreator(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"department": "search"})
	require.NoError(t, err)
	active := false

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{
			UserID: "admin",
			Roles:  []string{usecase.RoleSuperAdmin},
		}, nil, nil
	}
	mock.createUser = func(ctx context.Context, input usecase.CreateUserInput) (*model.UserView, error) {
		assert.Equal(t, "alice", input.Login)
		assert.Equal(t, "secret", input.Password)
		require.NotNil(t, input.IsActive)
		assert.Equal(t, active, *input.IsActive)
		require.NotNil(t, input.CreatedBy)
		assert.Equal(t, "admin", *input.CreatedBy)
		assert.Equal(t, "search", input.Attributes["department"])

		return &model.UserView{ID: "u1", Login: input.Login, IsActive: active, Roles: input.RoleCodes, Attributes: input.Attributes}, nil
	}

	resp, err := (&UserServiceServerImpl{svc: mock}).CreateUser(authContext("token"), &pb.CreateUserRequest{
		Login:      "alice",
		Password:   "secret",
		IsActive:   &active,
		RoleCodes:  []string{usecase.RoleUser},
		Attributes: attrs,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUser().GetId())
	assert.Equal(t, "alice", resp.GetUser().GetLogin())
}
