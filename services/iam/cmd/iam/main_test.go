package main

import (
	"context"
	"testing"

	transportgrpc "secure-rag-platform/services/iam/internal/transport/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMHealthRPC(t *testing.T) {
	t.Parallel()

	server := transportgrpc.NewIAMServiceServer(nil)

	resp, err := server.Health(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.GetStatus())
}
