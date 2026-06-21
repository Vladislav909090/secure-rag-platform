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

func TestAttributeServiceDeleteUserAttributeKeyUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}, nil, nil
	}
	mock.deleteUserAttributeKey = func(ctx context.Context, userID string, key string, updatedBy *string) (map[string]any, int64, error) {
		assert.Equal(t, "u1", userID)
		assert.Equal(t, "title", key)
		require.NotNil(t, updatedBy)
		assert.Equal(t, "admin", *updatedBy)

		return map[string]any{"department": "search"}, 12, nil
	}

	resp, err := (&AttributeServiceServerImpl{svc: mock}).DeleteUserAttributeKey(authContext("token"), &pb.DeleteUserAttributeKeyRequest{
		UserId: "u1",
		Key:    "title",
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(12), resp.GetCtxVer())
	assert.Equal(t, "search", resp.GetAttributes().AsMap()["department"])
}
