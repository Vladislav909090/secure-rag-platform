package grpc

import (
	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// KnowledgeServiceServerImpl реализует gRPC-сервис KnowledgeService
type KnowledgeServiceServerImpl struct {
	pb.UnimplementedKnowledgeServiceServer
	uc DocumentUsecaseContract
}

// NewKnowledgeServiceServer создаёт gRPC-сервер; при nil доступен только health
func NewKnowledgeServiceServer(uc *usecase.DocumentUsecase) *KnowledgeServiceServerImpl {
	return &KnowledgeServiceServerImpl{uc: usecaseOrNil(uc)}
}

func (s *KnowledgeServiceServerImpl) requireUC() error {
	if s.uc == nil {
		return status.Error(codes.Unavailable, "service not configured")
	}

	return nil
}

var _ pb.KnowledgeServiceServer = (*KnowledgeServiceServerImpl)(nil)
