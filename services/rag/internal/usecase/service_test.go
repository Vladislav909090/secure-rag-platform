package usecase

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRAGTestService(t *testing.T, repo RAGRepo) *Service {
	t.Helper()

	knowledge := NewMockKnowledgeServiceClient(t)
	knowledge.EXPECT().
		GetDocument(mock.Anything, mock.Anything).
		Return(&knowledgev1.GetDocumentResponse{
			Document: &knowledgev1.Document{
				Uuid:       "doc-1",
				MimeType:   "text/plain",
				StorageKey: "documents/doc-1/file",
			},
		}, nil).
		Maybe()

	return &Service{
		repo:       repo,
		storage:    NewMockObjectStorage(t),
		knowledge:  knowledge,
		embedding:  NewMockEmbeddingServiceClient(t),
		generation: NewMockGenerationServiceClient(t),
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
	assert.True(t, newRAGTestService(t, NewMockRAGRepo(t)).Ready())
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
