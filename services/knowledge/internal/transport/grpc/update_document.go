package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
)

func (s *KnowledgeServiceServerImpl) UpdateDocument(
	ctx context.Context,
	req *pb.UpdateDocumentRequest,
) (*pb.UpdateDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	var title, description *string
	if req.Title != nil {
		t := req.GetTitle()
		title = &t
	}
	if req.Description != nil {
		d := req.GetDescription()
		description = &d
	}

	doc, err := s.uc.UpdateDocument(ctx, req.GetDocumentUuid(), title, description)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateDocumentResponse{Document: documentToProto(doc)}, nil
}
