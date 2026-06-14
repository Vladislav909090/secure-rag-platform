package grpc

import (
	pb "secure-rag-platform/api/gen/go/rag/v1"
	"secure-rag-platform/services/rag/internal/usecase"
)

// Server реализует gRPC-сервис RAGService.
type Server struct {
	pb.UnimplementedRAGServiceServer
	uc *usecase.Service
}

// NewServer создаёт новый gRPC сервер RAG.
func NewServer(uc *usecase.Service) *Server {
	return &Server{uc: uc}
}

var _ pb.RAGServiceServer = (*Server)(nil)
