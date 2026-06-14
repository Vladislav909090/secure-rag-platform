package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
)

func (s *KnowledgeServiceServerImpl) RestoreDocument(
	ctx context.Context,
	req *pb.RestoreDocumentRequest,
) (*pb.RestoreDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	doc, err := s.uc.RestoreDocument(ctx, req.GetDocumentUuid())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RestoreDocumentResponse{Document: documentToProto(doc)}, nil
}
