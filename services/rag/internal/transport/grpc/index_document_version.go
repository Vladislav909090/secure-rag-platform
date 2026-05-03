package grpc

import (
	"context"

	pb "secure-rag-platform/services/rag/gen/v1"
	"secure-rag-platform/services/rag/internal/usecase"
)

func (s *Server) IndexDocumentVersion(
	ctx context.Context,
	req *pb.IndexDocumentVersionRequest,
) (*pb.IndexDocumentVersionResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.IndexDocumentVersion(ctx, usecase.IndexDocumentVersionRequest{
		DocumentUUID:        req.GetDocumentUuid(),
		VersionNumber:       req.GetVersionNumber(),
		EmbeddingModelAlias: req.GetEmbeddingModelAlias(),
		ChunkSize:           int(req.GetChunkSize()),
		ChunkOverlap:        int(req.GetChunkOverlap()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IndexDocumentVersionResponse{
		DocumentUuid:           result.DocumentUUID,
		VersionNumber:          result.VersionNumber,
		ChunkCount:             int32(result.ChunkCount),
		EmbeddingDimension:     result.EmbeddingDimension,
		ResolvedEmbeddingModel: result.ResolvedEmbeddingModel,
	}, nil
}
