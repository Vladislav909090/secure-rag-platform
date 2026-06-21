package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/iam/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMServiceHealth(t *testing.T) {
	t.Parallel()

	resp, err := (&IAMServiceServerImpl{}).Health(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.GetStatus())
}
