package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/rag/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRAGServiceServerImplHealth(t *testing.T) {
	t.Parallel()

	resp, err := (&RAGServiceServerImpl{}).Health(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "ok", resp.GetStatus())
}
