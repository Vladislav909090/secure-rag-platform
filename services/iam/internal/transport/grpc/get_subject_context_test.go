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

func TestInternalGetSubjectContextUsesAuthenticatedUserWhenRequestEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockIAMUsecase{t: t}
	mock.authenticateAccessToken = func(ctx context.Context, accessToken string) (*usecase.Principal, *model.SubjectContext, error) {
		assert.Equal(t, "self-token", accessToken)

		return &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}, nil, nil
	}
	mock.getSubjectContextByUserID = func(ctx context.Context, userID string) (*model.SubjectContext, error) {
		assert.Equal(t, "u1", userID)

		return &model.SubjectContext{
			UserID:   "u1",
			Login:    "alice",
			IsActive: true,
			Roles:    []string{usecase.RoleUser},
			CtxVer:   4,
		}, nil
	}

	resp, err := (&InternalIAMServiceServerImpl{svc: mock}).GetSubjectContext(
		authContext("self-token"),
		&pb.GetSubjectContextRequest{},
	)
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetSubject().GetUserId())
	assert.Equal(t, int64(4), resp.GetSubject().GetCtxVer())
}
