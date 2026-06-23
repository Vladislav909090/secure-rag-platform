package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/rag/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRAGServerDeleteDocumentIndexUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockRAGUsecase{
		t: t,
		deleteDocumentIndex: func(_ context.Context, documentUUID string) error {
			assert.Equal(t, "doc-1", documentUUID)

			return nil
		},
	}

	resp, err := (&Server{uc: mock}).DeleteDocumentIndex(context.Background(), &pb.DeleteDocumentIndexRequest{DocumentUuid: "doc-1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "doc-1", resp.GetDocumentUuid())
	assert.True(t, resp.GetDeleted())
}
