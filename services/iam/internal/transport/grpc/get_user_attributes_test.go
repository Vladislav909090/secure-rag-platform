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
)

func TestAttributeServiceGetUserAttributesDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		GetUserAttributes(mock.Anything, "u1").
		RunAndReturn(func(ctx context.Context, userID string) (map[string]any, int64, error) {
			assert.Equal(t, "u1", userID)

			return map[string]any{"department": "search"}, 4, nil
		})

	resp, err := (&AttributeServiceServerImpl{svc: uc}).GetUserAttributes(authContext("token"), &pb.GetUserAttributesRequest{})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(4), resp.GetCtxVer())
	assert.Equal(t, "search", resp.GetAttributes().AsMap()["department"])
}
