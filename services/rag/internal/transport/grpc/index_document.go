package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/rag/v1"
	"secure-rag-platform/services/rag/internal/usecase"
)

func (s *Server) IndexDocument(
	ctx context.Context,
	req *pb.IndexDocumentRequest,
) (*pb.IndexDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.IndexDocument(ctx, usecase.IndexDocumentRequest{
		DocumentUUID:        req.GetDocumentUuid(),
		EmbeddingModelAlias: req.GetEmbeddingModelAlias(),
		ChunkSize:           int(req.GetChunkSize()),
		ChunkOverlap:        int(req.GetChunkOverlap()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IndexDocumentResponse{
		DocumentUuid:           result.DocumentUUID,
		ChunkCount:             int32(result.ChunkCount),
		EmbeddingDimension:     result.EmbeddingDimension,
		ResolvedEmbeddingModel: result.ResolvedEmbeddingModel,
	}, nil
}
