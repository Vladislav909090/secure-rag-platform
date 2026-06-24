package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceHealth(t *testing.T) {
	t.Parallel()

	resp, err := (&KnowledgeServiceServerImpl{}).Health(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.GetStatus())
}
