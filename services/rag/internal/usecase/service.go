package usecase

import (
	"log/slog"

	aiinferencev1 "secure-rag-platform/services/ai-inference/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/rag/internal/repository"
	"secure-rag-platform/services/rag/internal/storage"
)

type Defaults struct {
	EmbeddingModelAlias  string
	GenerationModelAlias string
	ChunkSize            int
	ChunkOverlap         int
	TopK                 int32
}

// Service содержит бизнес-логику RAG.
type Service struct {
	repo       *repository.Repo
	storage    *storage.S3Storage
	knowledge  knowledgev1.KnowledgeServiceClient
	embedding  aiinferencev1.EmbeddingServiceClient
	generation aiinferencev1.GenerationServiceClient
	defaults   Defaults
	logger     *slog.Logger
}

// NewService создаёт RAG usecase.
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

// Ready проверяет минимальную конфигурацию.
func (s *Service) Ready() bool {
	return s != nil &&
		s.repo != nil &&
		s.storage != nil &&
		s.knowledge != nil &&
		s.embedding != nil &&
		s.generation != nil
}
