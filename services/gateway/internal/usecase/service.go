package usecase

import (
	"log/slog"

	iamv1 "secure-rag-platform/services/iam/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	ragv1 "secure-rag-platform/services/rag/gen/v1"
)

type Defaults struct {
	TopK                 int32
	EmbeddingModelAlias  string
	GenerationModelAlias string
}

// Service содержит бизнес-логику gateway.
type Service struct {
	rag           ragv1.RAGServiceClient
	knowledge     knowledgev1.KnowledgeServiceClient
	iam           iamv1.InternalIAMServiceClient
	auth          iamv1.AuthServiceClient
	policy        PolicyAuthorizer
	defaults      Defaults
	disableAuth   bool
	disableFilter bool
	logger        *slog.Logger
}

// NewService создаёт gateway usecase.
func NewService(
	rag ragv1.RAGServiceClient,
	knowledge knowledgev1.KnowledgeServiceClient,
	iam iamv1.InternalIAMServiceClient,
	auth iamv1.AuthServiceClient,
	policy PolicyAuthorizer,
	defaults Defaults,
	disableAuth bool,
	disableFilter bool,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		rag:           rag,
		knowledge:     knowledge,
		iam:           iam,
		auth:          auth,
		policy:        policy,
		defaults:      defaults,
		disableAuth:   disableAuth,
		disableFilter: disableFilter,
		logger:        logger,
	}
}

// Ready проверяет минимальную конфигурацию.
func (s *Service) Ready() bool {
	if s == nil || s.rag == nil || s.knowledge == nil {
		return false
	}
	if !s.disableAuth && s.iam == nil {
		return false
	}
	if !s.disableAuth && s.auth == nil {
		return false
	}
	return true
}
