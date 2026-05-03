package grpc

import (
	"context"

	pb "secure-rag-platform/services/rag/gen/v1"
	"secure-rag-platform/services/rag/internal/usecase"
)

func (s *Server) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.Query(ctx, usecase.QueryRequest{
		Query:                req.GetQuery(),
		TopK:                 req.GetTopK(),
		DocumentUUIDs:        req.GetDocumentUuids(),
		EmbeddingModelAlias:  req.GetEmbeddingModelAlias(),
		GenerationModelAlias: req.GetGenerationModelAlias(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	contexts := make([]*pb.QueryContext, 0, len(result.Contexts))
	for _, ctxItem := range result.Contexts {
		contexts = append(contexts, &pb.QueryContext{
			DocumentUuid:  ctxItem.DocumentUUID,
			VersionNumber: ctxItem.VersionNumber,
			ChunkIndex:    ctxItem.ChunkIndex,
			Text:          ctxItem.Text,
			Score:         ctxItem.Score,
		})
	}

	return &pb.QueryResponse{
		Answer:                  result.Answer,
		Contexts:                contexts,
		ResolvedEmbeddingModel:  result.ResolvedEmbeddingModel,
		ResolvedGenerationModel: result.ResolvedGenerationModel,
	}, nil
}
