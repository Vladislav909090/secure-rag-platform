package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceRefreshTokenUsesUsecase(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		RefreshToken(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input usecase.RefreshTokenInput) (*usecase.TokenPair, error) {
			assert.Equal(t, "old-refresh", input.RefreshToken)

			return &usecase.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
				ExpiresIn:    1800,
				TokenType:    "Bearer",
			}, nil
		})

	resp, err := (&AuthServiceServerImpl{svc: uc}).RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		RefreshToken: "old-refresh",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-access", resp.GetAccessToken())
	assert.Equal(t, "new-refresh", resp.GetRefreshToken())
	assert.Equal(t, int64(1800), resp.GetExpiresIn())
}
