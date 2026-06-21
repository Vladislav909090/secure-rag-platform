package main

import (
	"context"
	"testing"

	transportgrpc "secure-rag-platform/services/knowledge/internal/transport/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeHealthRPC(t *testing.T) {
	t.Parallel()

	server := &transportgrpc.KnowledgeServiceServerImpl{}

	resp, err := server.Health(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.GetStatus())
}
