package grpc

import (
	pb "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"
)

// KnowledgeServiceServerImpl реализует gRPC-сервис KnowledgeService.
type KnowledgeServiceServerImpl struct {
	pb.UnimplementedKnowledgeServiceServer
	uc *usecase.DocumentUsecase
}

// NewKnowledgeServiceServer создаёт новый gRPC-сервер. uc может быть nil (health всегда работает).
func NewKnowledgeServiceServer(uc *usecase.DocumentUsecase) *KnowledgeServiceServerImpl {
	return &KnowledgeServiceServerImpl{uc: uc}
}

var _ pb.KnowledgeServiceServer = (*KnowledgeServiceServerImpl)(nil)
