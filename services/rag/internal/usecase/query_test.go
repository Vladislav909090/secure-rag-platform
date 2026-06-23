package usecase

import (
	"context"
	"testing"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	"secure-rag-platform/services/rag/internal/repository"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestBuildPrompts(t *testing.T) {
	t.Parallel()

	system, user := buildPrompts("что внутри?", []QueryContext{
		{DocumentUUID: "doc-a", ChunkIndex: 2, Text: "первый факт"},
		{DocumentUUID: "doc-b", ChunkIndex: 1, Text: "второй факт"},
	})

	assert.NotEmpty(t, system)
	assert.Equal(t, "Вопрос:\nчто внутри?\n\nКонтекст:\n[doc-a:2]\nпервый факт\n\n[doc-b:1]\nвторой факт", user)
}

func TestRAGServiceQuerySearchesAndGeneratesAnswer(t *testing.T) {
	t.Parallel()

	repo := &mockRAGRepo{
		t: t,
		searchSimilar: func(
			_ context.Context,
			vector pgvector.Vector,
			embeddingModel string,
			embeddingDimension int32,
			indexedDimension int32,
			topK int32,
			documentUUIDs []string,
		) ([]repository.ChunkMatch, error) {
			assert.Equal(t, []float32{1, 2, 3, 4}, vector.Slice())
			assert.Equal(t, "embed-resolved", embeddingModel)
			assert.Equal(t, int32(4), embeddingDimension)
			assert.Equal(t, int32(4), indexedDimension)
			assert.Equal(t, int32(8), topK)
			assert.Equal(t, []string{"doc-1"}, documentUUIDs)

			return []repository.ChunkMatch{
				{DocumentUUID: "doc-1", ChunkIndex: 1, ChunkText: "first", Distance: 0.2},
				{DocumentUUID: "doc-1", ChunkIndex: 2, ChunkText: "far", Distance: 0.9},
			}, nil
		},
	}
	svc := newRAGTestService(t, repo)
	svc.embedding = &mockEmbeddingClient{
		t: t,
		embed: func(_ context.Context, req *aiinferencev1.EmbedRequest, _ ...grpc.CallOption) (*aiinferencev1.EmbedResponse, error) {
			assert.Equal(t, "embed-default", req.GetModelAlias())
			assert.Equal(t, "question", req.GetText())
			assert.True(t, req.GetNormalize())

			return &aiinferencev1.EmbedResponse{
				Vector:        []float32{1, 2, 3, 4},
				Dimension:     4,
				ResolvedModel: "embed-resolved",
			}, nil
		},
	}
	svc.generation = &mockGenerationClient{
		t: t,
		generate: func(_ context.Context, req *aiinferencev1.GenerateRequest, _ ...grpc.CallOption) (*aiinferencev1.GenerateResponse, error) {
			assert.Equal(t, "gen-default", req.GetModelAlias())
			require.Len(t, req.GetMessages(), 2)
			assert.Contains(t, req.GetMessages()[1].GetContent(), "first")
			assert.NotContains(t, req.GetMessages()[1].GetContent(), "far")

			return &aiinferencev1.GenerateResponse{Content: " answer ", ResolvedModel: "gen-resolved"}, nil
		},
	}

	got, err := svc.Query(context.Background(), QueryRequest{
		Query:         " question ",
		TopK:          2,
		DocumentUUIDs: []string{"doc-1"},
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "answer", got.Answer)
	assert.Equal(t, "embed-resolved", got.ResolvedEmbeddingModel)
	assert.Equal(t, "gen-resolved", got.ResolvedGenerationModel)
	require.Len(t, got.Contexts, 1)
	assert.Equal(t, float32(0.8), got.Contexts[0].Score)
}

func TestRAGServiceQueryReturnsFallbackWithoutContexts(t *testing.T) {
	t.Parallel()

	repo := &mockRAGRepo{
		t: t,
		searchSimilar: func(context.Context, pgvector.Vector, string, int32, int32, int32, []string) ([]repository.ChunkMatch, error) {
			return []repository.ChunkMatch{{DocumentUUID: "doc-1", ChunkIndex: 1, ChunkText: "far", Distance: 0.95}}, nil
		},
	}
	svc := newRAGTestService(t, repo)
	svc.embedding = &mockEmbeddingClient{
		t: t,
		embed: func(context.Context, *aiinferencev1.EmbedRequest, ...grpc.CallOption) (*aiinferencev1.EmbedResponse, error) {
			return &aiinferencev1.EmbedResponse{Vector: []float32{1, 2}, Dimension: 2, ResolvedModel: "embed"}, nil
		},
	}

	got, err := svc.Query(context.Background(), QueryRequest{Query: "question", DocumentUUIDs: []string{"doc-1"}})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "I don't know based on the provided documents.", got.Answer)
	assert.Empty(t, got.Contexts)
}

func TestRAGServiceQueryRejectsEmptyQuestion(t *testing.T) {
	t.Parallel()

	got, err := newRAGTestService(t, &mockRAGRepo{t: t}).Query(context.Background(), QueryRequest{Query: " "})
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, got)
}

func TestRAGServiceQueryRejectsEmbeddingDimensionMismatch(t *testing.T) {
	t.Parallel()

	svc := newRAGTestService(t, &mockRAGRepo{t: t})
	svc.embedding = &mockEmbeddingClient{
		t: t,
		embed: func(context.Context, *aiinferencev1.EmbedRequest, ...grpc.CallOption) (*aiinferencev1.EmbedResponse, error) {
			return &aiinferencev1.EmbedResponse{Vector: []float32{1, 2}, Dimension: 3}, nil
		},
	}

	got, err := svc.Query(context.Background(), QueryRequest{Query: "question"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embedding dimension mismatch")
	assert.Nil(t, got)
}
