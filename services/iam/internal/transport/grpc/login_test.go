package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{
		t: t,
		login: func(ctx context.Context, input usecase.LoginInput) (*usecase.TokenPair, error) {
			assert.Equal(t, "alice", input.Login)
			assert.Equal(t, "secret", input.Password)

			return &usecase.TokenPair{
				AccessToken:  "access",
				RefreshToken: "refresh",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			}, nil
		},
	}

	resp, err := (&AuthServiceServerImpl{svc: mock}).Login(context.Background(), &pb.LoginRequest{
		Login:    "alice",
		Password: "secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "access", resp.GetAccessToken())
	assert.Equal(t, "refresh", resp.GetRefreshToken())
	assert.Equal(t, "Bearer", resp.GetTokenType())
}
