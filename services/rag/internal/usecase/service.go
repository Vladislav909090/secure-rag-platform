package usecase

import (
	"context"
	"io"
	"log/slog"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/rag/internal/repository"
	"secure-rag-platform/services/rag/internal/storage"

	"github.com/pgvector/pgvector-go"
)

type Defaults struct {
	EmbeddingModelAlias  string
	GenerationModelAlias string
	ChunkSize            int
	ChunkOverlap         int
	TopK                 int32
	IndexedEmbeddingDim  int32
}

// Service содержит бизнес-логику RAG
type Service struct {
	repo       ragRepo
	storage    objectStorage
	knowledge  knowledgev1.KnowledgeServiceClient
	embedding  aiinferencev1.EmbeddingServiceClient
	generation aiinferencev1.GenerationServiceClient
	defaults   Defaults
	logger     *slog.Logger
}

type ragRepo interface {
	DeleteChunks(context.Context, string) error
	InsertChunks(context.Context, []repository.Chunk) error
	SearchSimilar(context.Context, pgvector.Vector, string, int32, int32, int32, []string) ([]repository.ChunkMatch, error)
}

type objectStorage interface {
	Download(context.Context, string) (io.ReadCloser, error)
}

// NewService создаёт сервисный слой RAG
func NewService(
	repo *repository.Repo,
	storage *storage.S3Storage,
	knowledge knowledgev1.KnowledgeServiceClient,
	embedding aiinferencev1.EmbeddingServiceClient,
	generation aiinferencev1.GenerationServiceClient,
	defaults Defaults,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		repo:       repo,
		storage:    storage,
		knowledge:  knowledge,
		embedding:  embedding,
		generation: generation,
		defaults:   defaults,
		logger:     logger,
	}
}

// Ready проверяет минимальную конфигурацию
func (s *Service) Ready() bool {
	return s != nil &&
		s.repo != nil &&
		s.storage != nil &&
		s.knowledge != nil &&
		s.embedding != nil &&
		s.generation != nil
}
