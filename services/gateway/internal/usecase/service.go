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
	users         iamv1.UserServiceClient
	roles         iamv1.RoleServiceClient
	attributes    iamv1.AttributeServiceClient
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
	users iamv1.UserServiceClient,
	roles iamv1.RoleServiceClient,
	attributes iamv1.AttributeServiceClient,
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
		users:         users,
		roles:         roles,
		attributes:    attributes,
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
	if s.disableAuth {
		return true
	}
	return s.iam != nil &&
		s.auth != nil &&
		s.users != nil &&
		s.roles != nil &&
		s.attributes != nil
}
