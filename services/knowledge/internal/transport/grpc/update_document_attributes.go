package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) UpdateDocumentAttributes(
	ctx context.Context,
	req *pb.UpdateDocumentAttributesRequest,
) (*pb.UpdateDocumentAttributesResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	var attrs map[string]any
	if req.GetAttributes() != nil {
		attrs = req.GetAttributes().AsMap()
	}

	doc, err := s.uc.UpdateAttributes(ctx, req.GetDocumentUuid(), attrs)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateDocumentAttributesResponse{Document: documentToProto(doc)}, nil
}
