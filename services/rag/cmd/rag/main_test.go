package main

import (
	"context"
	"testing"

	transportgrpc "secure-rag-platform/services/rag/internal/transport/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRAGHealthRPC(t *testing.T) {
	t.Parallel()

	server := &transportgrpc.RAGServiceServerImpl{}

	resp, err := server.Health(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.GetStatus())
}

func TestParseInt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10, parseInt("10", 3))
	assert.Equal(t, 3, parseInt("", 3))
	assert.Equal(t, 3, parseInt("bad", 3))
	assert.Equal(t, 3, parseInt("-1", 3))
}
