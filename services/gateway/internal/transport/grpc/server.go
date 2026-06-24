package grpc

import (
	pb "secure-rag-platform/api/gen/go/gateway/v1"
	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC-сервисы Gateway
type Server struct {
	pb.UnimplementedGatewayServiceServer
	pb.UnimplementedGatewayAuthServiceServer
	pb.UnimplementedGatewayKnowledgeServiceServer
	pb.UnimplementedGatewayRAGServiceServer
	uc *usecase.Service
}

// NewServer создаёт новый gRPC сервер gateway
func NewServer(uc *usecase.Service) *Server {
	return &Server{uc: uc}
}

func (s *Server) requireUC() error {
	if s.uc == nil || !s.uc.Ready() {
		return status.Error(codes.Unavailable, "service not configured")
	}

	return nil
}

var _ pb.GatewayServiceServer = (*Server)(nil)
var _ pb.GatewayAuthServiceServer = (*Server)(nil)
var _ pb.GatewayKnowledgeServiceServer = (*Server)(nil)
var _ pb.GatewayRAGServiceServer = (*Server)(nil)
