package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/rag/v1"
	"secure-rag-platform/services/rag/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRAGServerIndexDocumentMapsRequestAndResult(t *testing.T) {
	t.Parallel()

	mock := &mockRAGUsecase{
		t: t,
		indexDocument: func(_ context.Context, req usecase.IndexDocumentRequest) (*usecase.IndexDocumentResult, error) {
			assert.Equal(t, "doc-1", req.DocumentUUID)
			assert.Equal(t, "embed", req.EmbeddingModelAlias)
			assert.Equal(t, 256, req.ChunkSize)
			assert.Equal(t, 32, req.ChunkOverlap)

			return &usecase.IndexDocumentResult{
				DocumentUUID:           "doc-1",
				ChunkCount:             3,
				EmbeddingDimension:     4,
				ResolvedEmbeddingModel: "embed-model",
			}, nil
		},
	}

	resp, err := (&Server{uc: mock}).IndexDocument(context.Background(), &pb.IndexDocumentRequest{
		DocumentUuid:        "doc-1",
		EmbeddingModelAlias: "embed",
		ChunkSize:           256,
		ChunkOverlap:        32,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "doc-1", resp.GetDocumentUuid())
	assert.Equal(t, int32(3), resp.GetChunkCount())
	assert.Equal(t, int32(4), resp.GetEmbeddingDimension())
	assert.Equal(t, "embed-model", resp.GetResolvedEmbeddingModel())
}
