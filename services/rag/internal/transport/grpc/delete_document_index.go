package grpc

import (
	"context"

	pb "secure-rag-platform/services/rag/gen/v1"
)

func (s *Server) DeleteDocumentIndex(
	ctx context.Context,
	req *pb.DeleteDocumentIndexRequest,
) (*pb.DeleteDocumentIndexResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	if err := s.uc.DeleteDocumentIndex(ctx, req.GetDocumentUuid()); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteDocumentIndexResponse{
		DocumentUuid: req.GetDocumentUuid(),
		Deleted:      true,
	}, nil
}
