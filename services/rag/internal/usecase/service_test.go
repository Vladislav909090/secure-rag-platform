package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/rag/internal/repository"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockRAGRepo struct {
	t *testing.T

	deleteChunks  func(context.Context, string) error
	insertChunks  func(context.Context, []repository.Chunk) error
	searchSimilar func(context.Context, pgvector.Vector, string, int32, int32, int32, []string) ([]repository.ChunkMatch, error)
}

var _ ragRepo = (*mockRAGRepo)(nil)

func (m *mockRAGRepo) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected repo call", "unexpected repo call: %s", name)
	}
	panic(fmt.Sprintf("unexpected repo call: %s", name))
}

func (m *mockRAGRepo) DeleteChunks(ctx context.Context, documentUUID string) error {
	if m.deleteChunks == nil {
		m.unexpected("DeleteChunks")
	}

	return m.deleteChunks(ctx, documentUUID)
}

func (m *mockRAGRepo) InsertChunks(ctx context.Context, chunks []repository.Chunk) error {
	if m.insertChunks == nil {
		m.unexpected("InsertChunks")
	}

	return m.insertChunks(ctx, chunks)
}

func (m *mockRAGRepo) SearchSimilar(
	ctx context.Context,
	vector pgvector.Vector,
	embeddingModel string,
	embeddingDimension int32,
	indexedDimension int32,
	topK int32,
	documentUUIDs []string,
) ([]repository.ChunkMatch, error) {
	if m.searchSimilar == nil {
		m.unexpected("SearchSimilar")
	}

	return m.searchSimilar(ctx, vector, embeddingModel, embeddingDimension, indexedDimension, topK, documentUUIDs)
}

type mockObjectStorage struct {
	t *testing.T

	download func(context.Context, string) (io.ReadCloser, error)
}

var _ objectStorage = (*mockObjectStorage)(nil)

func (m *mockObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.download == nil {
		if m.t != nil {
			m.t.Helper()
			require.FailNowf(m.t, "unexpected storage call", "unexpected storage call: %s", "Download")
		}
		panic("unexpected storage call: Download")
	}

	return m.download(ctx, key)
}

type mockKnowledgeClient struct {
	knowledgev1.KnowledgeServiceClient
	t *testing.T

	getDocument func(context.Context, *knowledgev1.GetDocumentRequest, ...grpc.CallOption) (*knowledgev1.GetDocumentResponse, error)
}

func (m *mockKnowledgeClient) GetDocument(ctx context.Context, req *knowledgev1.GetDocumentRequest, opts ...grpc.CallOption) (*knowledgev1.GetDocumentResponse, error) {
	if m.getDocument == nil {
		m.t.Helper()
		require.FailNow(m.t, "unexpected knowledge call: GetDocument")
	}

	return m.getDocument(ctx, req, opts...)
}

type mockEmbeddingClient struct {
	aiinferencev1.EmbeddingServiceClient
	t *testing.T

	embed      func(context.Context, *aiinferencev1.EmbedRequest, ...grpc.CallOption) (*aiinferencev1.EmbedResponse, error)
	batchEmbed func(context.Context, *aiinferencev1.BatchEmbedRequest, ...grpc.CallOption) (*aiinferencev1.BatchEmbedResponse, error)
}

func (m *mockEmbeddingClient) Embed(ctx context.Context, req *aiinferencev1.EmbedRequest, opts ...grpc.CallOption) (*aiinferencev1.EmbedResponse, error) {
	if m.embed == nil {
		m.t.Helper()
		require.FailNow(m.t, "unexpected embedding call: Embed")
	}

	return m.embed(ctx, req, opts...)
}

func (m *mockEmbeddingClient) BatchEmbed(ctx context.Context, req *aiinferencev1.BatchEmbedRequest, opts ...grpc.CallOption) (*aiinferencev1.BatchEmbedResponse, error) {
	if m.batchEmbed == nil {
		m.t.Helper()
		require.FailNow(m.t, "unexpected embedding call: BatchEmbed")
	}

	return m.batchEmbed(ctx, req, opts...)
}

type mockGenerationClient struct {
	aiinferencev1.GenerationServiceClient
	t *testing.T

	generate func(context.Context, *aiinferencev1.GenerateRequest, ...grpc.CallOption) (*aiinferencev1.GenerateResponse, error)
}

func (m *mockGenerationClient) Generate(ctx context.Context, req *aiinferencev1.GenerateRequest, opts ...grpc.CallOption) (*aiinferencev1.GenerateResponse, error) {
	if m.generate == nil {
		m.t.Helper()
		require.FailNow(m.t, "unexpected generation call: Generate")
	}

	return m.generate(ctx, req, opts...)
}

func newRAGTestService(t *testing.T, repo ragRepo) *Service {
	t.Helper()

	return &Service{
		repo:    repo,
		storage: &mockObjectStorage{t: t},
		knowledge: &mockKnowledgeClient{
			t: t,
			getDocument: func(context.Context, *knowledgev1.GetDocumentRequest, ...grpc.CallOption) (*knowledgev1.GetDocumentResponse, error) {
				return &knowledgev1.GetDocumentResponse{
					Document: &knowledgev1.Document{
						Uuid:       "doc-1",
						MimeType:   "text/plain",
						StorageKey: "documents/doc-1/file",
					},
				}, nil
			},
		},
		embedding:  &mockEmbeddingClient{t: t},
		generation: &mockGenerationClient{t: t},
		defaults: Defaults{
			EmbeddingModelAlias:  "embed-default",
			GenerationModelAlias: "gen-default",
			ChunkSize:            64,
			ChunkOverlap:         8,
			TopK:                 3,
			IndexedEmbeddingDim:  4,
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestRAGServiceReady(t *testing.T) {
	t.Parallel()

	assert.False(t, (*Service)(nil).Ready())
	assert.False(t, (&Service{}).Ready())
	assert.True(t, newRAGTestService(t, &mockRAGRepo{t: t}).Ready())
}

func TestNewServiceSetsLoggerAndDependencies(t *testing.T) {
	t.Parallel()

	svc := NewService(nil, nil, nil, nil, nil, Defaults{TopK: 5}, nil)
	require.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.Equal(t, int32(5), svc.defaults.TopK)
	assert.False(t, svc.Ready())
}

func readCloser(text string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(text))
}
