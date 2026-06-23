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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAttributeServiceReplaceUserAttributesMapsRequest(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"team": "rag"})
	require.NoError(t, err)
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(&usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		ReplaceUserAttributes(mock.Anything, "u1", mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string, gotAttrs map[string]any, updatedBy *string) (map[string]any, int64, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "rag", gotAttrs["team"])
			require.NotNil(t, updatedBy)
			assert.Equal(t, "admin", *updatedBy)

			return gotAttrs, 10, nil
		})

	resp, err := (&AttributeServiceServerImpl{svc: uc}).ReplaceUserAttributes(authContext("token"), &pb.ReplaceUserAttributesRequest{
		UserId:     "u1",
		Attributes: attrs,
	})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	assert.Equal(t, int64(10), resp.GetCtxVer())
	assert.Equal(t, "rag", resp.GetAttributes().AsMap()["team"])
}
