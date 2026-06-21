package grpc

import (
	pb "secure-rag-platform/api/gen/go/rag/v1"
	"secure-rag-platform/services/rag/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC-сервис RAGService
type Server struct {
	pb.UnimplementedRAGServiceServer
	uc *usecase.Service
}

// NewServer создаёт новый gRPC сервер RAG
func NewServer(uc *usecase.Service) *Server {
	return &Server{uc: uc}
}

func (s *Server) requireUC() error {
	if s.uc == nil || !s.uc.Ready() {
		return status.Error(codes.Unavailable, "service not configured")
	}

	return nil
}

var _ pb.RAGServiceServer = (*Server)(nil)
