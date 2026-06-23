package grpc

import (
	"context"
	"testing"
	"time"

	pb "secure-rag-platform/api/gen/go/iam/v1"
	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInternalValidateAccessTokenFallsBackToMetadata(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		ValidateAccessToken(mock.Anything, "fallback-token").
		RunAndReturn(func(ctx context.Context, accessToken string) (*usecase.ValidateTokenResult, error) {
			assert.Equal(t, "fallback-token", accessToken)

			return &usecase.ValidateTokenResult{
				Valid: true,
				Subject: &model.SubjectContext{
					UserID:   "u1",
					Login:    "alice",
					IsActive: true,
					Roles:    []string{usecase.RoleUser},
					CtxVer:   3,
				},
				Principal: &usecase.Principal{
					SessionID: "s1",
					TokenID:   "jti1",
					ExpiresAt: expiresAt,
				},
				ExpiresAtUnix: expiresAt.Unix(),
			}, nil
		})

	resp, err := (&InternalIAMServiceServerImpl{svc: uc}).ValidateAccessToken(
		authContext("fallback-token"),
		&pb.ValidateAccessTokenRequest{},
	)
	require.NoError(t, err)
	assert.True(t, resp.GetValid())
	assert.Equal(t, "u1", resp.GetSubject().GetUserId())
	assert.Equal(t, "s1", resp.GetSessionId())
	assert.Equal(t, "jti1", resp.GetTokenId())
}
