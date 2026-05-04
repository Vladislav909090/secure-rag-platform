package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) ReindexDocument(
	ctx context.Context,
	req *pb.ReindexDocumentRequest,
) (*pb.ReindexDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	result, err := s.uc.ReindexDocument(ctx, req.GetDocumentUuid())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ReindexDocumentResponse{
		DocumentUuid: result.DocumentUUID,
		IndexStatus:  result.IndexStatus,
	}, nil
}
