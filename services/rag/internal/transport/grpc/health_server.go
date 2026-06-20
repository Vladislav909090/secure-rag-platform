package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/rag/v1"
)

func (s *Server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "ok"}, nil
}
