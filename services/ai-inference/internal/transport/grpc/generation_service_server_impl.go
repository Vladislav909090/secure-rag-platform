package grpc

import (
	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

type GenerationServiceServerImpl struct {
	aiinferencev1.UnimplementedGenerationServiceServer
	svc *usecase.Service
}

func NewGenerationServiceServer(svc *usecase.Service) *GenerationServiceServerImpl {
	return &GenerationServiceServerImpl{svc: svc}
}

var _ aiinferencev1.GenerationServiceServer = (*GenerationServiceServerImpl)(nil)
