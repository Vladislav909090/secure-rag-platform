package usecase

import (
	"context"
	"io"
	"testing"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/rag/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestIsTextMime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  bool
	}{
		{"", true},
		{"text/plain; charset=utf-8", true},
		{"application/json", true},
		{"application/xml", true},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isTextMime(tt.value))
		})
	}
}

func TestRAGServiceIndexDocumentDownloadsEmbedsAndStoresChunks(t *testing.T) {
	t.Parallel()

	repo := &mockRAGRepo{
		t: t,
		deleteChunks: func(_ context.Context, documentUUID string) error {
			assert.Equal(t, "doc-1", documentUUID)

			return nil
		},
		insertChunks: func(_ context.Context, chunks []repository.Chunk) error {
			require.Len(t, chunks, 6)
			assert.Equal(t, "doc-1", chunks[0].DocumentUUID)
			assert.Equal(t, int32(0), chunks[0].ChunkIndex)
			assert.Equal(t, "embed-resolved", chunks[0].EmbeddingModel)
			assert.Equal(t, int32(2), chunks[0].EmbeddingDimension)
			assert.Equal(t, []float32{0, 1}, chunks[0].Embedding.Slice())
			assert.Equal(t, int32(5), chunks[5].ChunkIndex)
			assert.Equal(t, []float32{5, 6}, chunks[5].Embedding.Slice())

			return nil
		},
	}
	svc := newRAGTestService(t, repo)
	svc.storage = &mockObjectStorage{
		t: t,
		download: func(_ context.Context, key string) (io.ReadCloser, error) {
			assert.Equal(t, "documents/doc-1/file", key)

			return readCloser("abcdefghijklmnopqrst"), nil
		},
	}
	svc.embedding = &mockEmbeddingClient{
		t: t,
		batchEmbed: func(_ context.Context, req *aiinferencev1.BatchEmbedRequest, _ ...grpc.CallOption) (*aiinferencev1.BatchEmbedResponse, error) {
			assert.Equal(t, "embed-default", req.GetModelAlias())
			assert.True(t, req.GetNormalize())
			require.Len(t, req.GetTexts(), 6)
			assert.Equal(t, "abcdefghij", req.GetTexts()[0])
			assert.Equal(t, "klmnopqrst", req.GetTexts()[5])

			vectors := make([]*aiinferencev1.EmbeddingVector, 0, len(req.GetTexts()))
			for i := range req.GetTexts() {
				vectors = append(vectors, &aiinferencev1.EmbeddingVector{Values: []float32{float32(i), float32(i + 1)}})
			}

			return &aiinferencev1.BatchEmbedResponse{
				Vectors:       vectors,
				Dimension:     2,
				ResolvedModel: "embed-resolved",
			}, nil
		},
	}

	got, err := svc.IndexDocument(context.Background(), IndexDocumentRequest{
		DocumentUUID: " doc-1 ",
		ChunkSize:    10,
		ChunkOverlap: 0,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "doc-1", got.DocumentUUID)
	assert.Equal(t, 6, got.ChunkCount)
	assert.Equal(t, int32(2), got.EmbeddingDimension)
	assert.Equal(t, "embed-resolved", got.ResolvedEmbeddingModel)
}

func TestRAGServiceIndexDocumentRejectsEmptyDocumentUUID(t *testing.T) {
	t.Parallel()

	got, err := newRAGTestService(t, &mockRAGRepo{t: t}).IndexDocument(context.Background(), IndexDocumentRequest{})
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, got)
}

func TestRAGServiceIndexDocumentRejectsUnsupportedMime(t *testing.T) {
	t.Parallel()

	svc := newRAGTestService(t, &mockRAGRepo{t: t})
	svc.knowledge = &mockKnowledgeClient{
		t: t,
		getDocument: func(context.Context, *knowledgev1.GetDocumentRequest, ...grpc.CallOption) (*knowledgev1.GetDocumentResponse, error) {
			return &knowledgev1.GetDocumentResponse{
				Document: &knowledgev1.Document{Uuid: "doc-1", StorageKey: "key", MimeType: "application/pdf"},
			}, nil
		},
	}

	got, err := svc.IndexDocument(context.Background(), IndexDocumentRequest{DocumentUUID: "doc-1"})
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, got)
}

func TestRAGServiceIndexDocumentRejectsEmbeddingSizeMismatch(t *testing.T) {
	t.Parallel()

	svc := newRAGTestService(t, &mockRAGRepo{t: t})
	svc.storage = &mockObjectStorage{
		t:        t,
		download: func(context.Context, string) (io.ReadCloser, error) { return readCloser("alpha beta"), nil },
	}
	svc.embedding = &mockEmbeddingClient{
		t: t,
		batchEmbed: func(context.Context, *aiinferencev1.BatchEmbedRequest, ...grpc.CallOption) (*aiinferencev1.BatchEmbedResponse, error) {
			return &aiinferencev1.BatchEmbedResponse{Vectors: nil, Dimension: 2}, nil
		},
	}

	got, err := svc.IndexDocument(context.Background(), IndexDocumentRequest{DocumentUUID: "doc-1", ChunkSize: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embedding size mismatch")
	assert.Nil(t, got)
}
