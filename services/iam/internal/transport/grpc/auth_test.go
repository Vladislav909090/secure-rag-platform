package grpc

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestIAMExtractBearerToken(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", " Bearer token "))
	token, err := extractBearerToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "token", token)

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "token"))
	_, err = extractBearerToken(ctx)
	require.ErrorIs(t, err, usecase.ErrUnauthorized)
}

func TestIAMAccessHelpers(t *testing.T) {
	t.Parallel()

	principal := &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}
	assert.True(t, canAccessUser(principal, "u1"))
	assert.False(t, canAccessUser(principal, "u2"))

	admin := &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleAccessAdmin}}
	assert.True(t, canAccessUser(admin, "u2"))
	assert.True(t, isAdmin(admin))
}

func TestIAMRequireUC(t *testing.T) {
	t.Parallel()

	assert.Equal(t, codes.Unavailable, status.Code(requireUC(nil)))
	require.NoError(t, requireUC(usecase.NewIAMUsecase(nil, nil, usecase.DefaultConfig(), nil)))
}
