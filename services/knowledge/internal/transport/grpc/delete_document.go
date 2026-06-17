package grpc

import (
	"context"
	"time"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
)

func (s *KnowledgeServiceServerImpl) DeleteDocument(
	ctx context.Context,
	req *pb.DeleteDocumentRequest,
) (*pb.DeleteDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	result, err := s.uc.DeleteDocument(ctx, req.GetDocumentUuid())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteDocumentResponse{
		DocumentUuid: result.DocumentUUID,
		Deleted:      true,
		DeletedAt:    result.DeletedAt.Format(time.RFC3339),
	}, nil
}
