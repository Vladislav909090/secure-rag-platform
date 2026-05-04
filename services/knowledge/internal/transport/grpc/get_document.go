package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) GetDocument(
	ctx context.Context,
	req *pb.GetDocumentRequest,
) (*pb.GetDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	result, err := s.uc.GetDocument(ctx, req.GetDocumentUuid())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetDocumentResponse{
		Document: documentToProto(result.Document),
	}, nil
}
