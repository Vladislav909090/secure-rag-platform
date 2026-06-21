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

func TestAttributeServicePatchUserAttributesDefaultsToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	patch, err := structpb.NewStruct(map[string]any{"title": "editor"})
	require.NoError(t, err)
	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
	}
	mock.patchUserAttributes = func(ctx context.Context, userID string, gotPatch map[string]any, updatedBy *string) (map[string]any, int64, error) {
		assert.Equal(t, "u1", userID)
		assert.Equal(t, "editor", gotPatch["title"])
		require.NotNil(t, updatedBy)
		assert.Equal(t, "u1", *updatedBy)

		return map[string]any{"title": "editor", "department": "search"}, 11, nil
	}

	resp, err := (&AttributeServiceServerImpl{svc: mock}).PatchUserAttributes(authContext("token"), &pb.PatchUserAttributesRequest{Attributes: patch})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(11), resp.GetCtxVer())
	assert.Equal(t, "editor", resp.GetAttributes().AsMap()["title"])
}
