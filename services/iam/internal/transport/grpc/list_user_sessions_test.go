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

func TestSessionServiceListUserSessionsUsesPrincipal(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "admin", Roles: []string{usecase.RoleAccessAdmin}}
	expiresAt := time.Date(2026, 6, 22, 13, 0, 0, 0, time.UTC)
	uc := NewMockIAMUsecase(t)
	uc.EXPECT().
		AuthenticateAccessToken(mock.Anything, "token").
		Return(principal, (*model.SubjectContext)(nil), nil)
	uc.EXPECT().
		ListUserSessions(mock.Anything, principal, "u1").
		RunAndReturn(func(ctx context.Context, got *usecase.Principal, userID string) ([]*model.UserSession, string, error) {
			assert.Same(t, principal, got)
			assert.Equal(t, "u1", userID)

			return []*model.UserSession{{ID: "s1", UserID: userID, ExpiresAt: expiresAt}}, userID, nil
		})

	resp, err := (&SessionServiceServerImpl{svc: uc}).ListUserSessions(authContext("token"), &pb.ListUserSessionsRequest{UserId: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "u1", resp.GetUserId())
	require.Len(t, resp.GetSessions(), 1)
	assert.Equal(t, "s1", resp.GetSessions()[0].GetId())
}
