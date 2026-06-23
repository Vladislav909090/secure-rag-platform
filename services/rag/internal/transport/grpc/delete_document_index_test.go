package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/rag/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRAGServerDeleteDocumentIndexUsesUsecase(t *testing.T) {
	t.Parallel()

	uc := NewMockRAGUsecase(t)
	uc.EXPECT().Ready().Return(true)
	uc.EXPECT().
		DeleteDocumentIndex(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, documentUUID string) error {
			assert.Equal(t, "doc-1", documentUUID)

			return nil
		})

	resp, err := (&Server{uc: uc}).DeleteDocumentIndex(context.Background(), &pb.DeleteDocumentIndexRequest{DocumentUuid: "doc-1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "doc-1", resp.GetDocumentUuid())
	assert.True(t, resp.GetDeleted())
}
