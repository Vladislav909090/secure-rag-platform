package grpc

import (
	"context"

	pb "secure-rag-platform/services/gateway/gen/v1"
	"secure-rag-platform/services/gateway/internal/usecase"
)

func (s *Server) ReindexDocument(
	ctx context.Context,
	req *pb.ReindexDocumentRequest,
) (*pb.ReindexDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.ReindexDocument(ctx, usecase.ReindexRequest{
		DocumentUUID:        req.GetDocumentUuid(),
		EmbeddingModelAlias: req.GetEmbeddingModelAlias(),
		ChunkSize:           req.GetChunkSize(),
		ChunkOverlap:        req.GetChunkOverlap(),
	}, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ReindexDocumentResponse{
		DocumentUuid:           result.DocumentUUID,
		ChunkCount:             result.ChunkCount,
		EmbeddingDimension:     result.EmbeddingDimension,
		ResolvedEmbeddingModel: result.ResolvedEmbeddingModel,
	}, nil
}

func (s *Server) ReindexAllDocuments(
	ctx context.Context,
	req *pb.ReindexAllDocumentsRequest,
) (*pb.ReindexAllDocumentsResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.ReindexAllDocuments(ctx, usecase.ReindexRequest{
		EmbeddingModelAlias: req.GetEmbeddingModelAlias(),
		ChunkSize:           req.GetChunkSize(),
		ChunkOverlap:        req.GetChunkOverlap(),
	}, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	items := make([]*pb.ReindexDocumentResult, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &pb.ReindexDocumentResult{
			DocumentUuid:           item.DocumentUUID,
			Indexed:                item.Indexed,
			Error:                  item.Error,
			ChunkCount:             item.ChunkCount,
			EmbeddingDimension:     item.EmbeddingDimension,
			ResolvedEmbeddingModel: item.ResolvedEmbeddingModel,
		})
	}

	return &pb.ReindexAllDocumentsResponse{
		TotalCount:   result.TotalCount,
		IndexedCount: result.IndexedCount,
		FailedCount:  result.FailedCount,
		Items:        items,
	}, nil
}
