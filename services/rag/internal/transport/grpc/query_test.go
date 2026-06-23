package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/rag/v1"
	"secure-rag-platform/services/rag/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRAGServerQueryMapsRequestAndResult(t *testing.T) {
	t.Parallel()

	mock := &mockRAGUsecase{
		t: t,
		query: func(_ context.Context, req usecase.QueryRequest) (*usecase.QueryResult, error) {
			assert.Equal(t, "question", req.Query)
			assert.Equal(t, int32(2), req.TopK)
			assert.Equal(t, []string{"doc-1", "doc-2"}, req.DocumentUUIDs)
			assert.Equal(t, "embed", req.EmbeddingModelAlias)
			assert.Equal(t, "gen", req.GenerationModelAlias)

			return &usecase.QueryResult{
				Answer: "answer",
				Contexts: []usecase.QueryContext{
					{DocumentUUID: "doc-1", ChunkIndex: 1, Text: "context", Score: 0.75},
				},
				ResolvedEmbeddingModel:  "embed-model",
				ResolvedGenerationModel: "gen-model",
			}, nil
		},
	}

	resp, err := (&Server{uc: mock}).Query(context.Background(), &pb.QueryRequest{
		Query:                "question",
		TopK:                 2,
		DocumentUuids:        []string{"doc-1", "doc-2"},
		EmbeddingModelAlias:  "embed",
		GenerationModelAlias: "gen",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "answer", resp.GetAnswer())
	assert.Equal(t, "embed-model", resp.GetResolvedEmbeddingModel())
	assert.Equal(t, "gen-model", resp.GetResolvedGenerationModel())
	require.Len(t, resp.GetContexts(), 1)
	assert.Equal(t, "doc-1", resp.GetContexts()[0].GetDocumentUuid())
	assert.Equal(t, int32(1), resp.GetContexts()[0].GetChunkIndex())
	assert.Equal(t, "context", resp.GetContexts()[0].GetText())
	assert.InDelta(t, 0.75, resp.GetContexts()[0].GetScore(), 0.001)
}
