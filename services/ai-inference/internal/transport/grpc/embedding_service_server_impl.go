package grpc

import (
	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

type EmbeddingServiceServerImpl struct {
	aiinferencev1.UnimplementedEmbeddingServiceServer
	svc *usecase.Service
}

func NewEmbeddingServiceServer(svc *usecase.Service) *EmbeddingServiceServerImpl {
	return &EmbeddingServiceServerImpl{svc: svc}
}

var _ aiinferencev1.EmbeddingServiceServer = (*EmbeddingServiceServerImpl)(nil)
