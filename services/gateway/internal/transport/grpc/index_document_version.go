package grpc

import (
	"context"

	pb "secure-rag-platform/services/gateway/gen/v1"
	"secure-rag-platform/services/gateway/internal/usecase"
)

func (s *Server) IndexDocumentVersion(ctx context.Context, req *pb.IndexDocumentVersionRequest) (*pb.IndexDocumentVersionResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.IndexDocumentVersion(ctx, usecase.IndexRequest{
		DocumentUUID:        req.GetDocumentUuid(),
		VersionNumber:       req.GetVersionNumber(),
		EmbeddingModelAlias: req.GetEmbeddingModelAlias(),
		ChunkSize:           req.GetChunkSize(),
		ChunkOverlap:        req.GetChunkOverlap(),
	}, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IndexDocumentVersionResponse{
		DocumentUuid:           result.DocumentUUID,
		VersionNumber:          result.VersionNumber,
		ChunkCount:             result.ChunkCount,
		EmbeddingDimension:     result.EmbeddingDimension,
		ResolvedEmbeddingModel: result.ResolvedEmbeddingModel,
	}, nil
}
