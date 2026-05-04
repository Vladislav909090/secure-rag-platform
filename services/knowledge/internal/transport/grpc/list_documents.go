package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) ListDocuments(
	ctx context.Context,
	_ *pb.ListDocumentsRequest,
) (*pb.ListDocumentsResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	items, err := s.uc.ListDocuments(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbItems := make([]*pb.DocumentItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &pb.DocumentItem{
			Document: documentToProto(item.Document),
		})
	}

	return &pb.ListDocumentsResponse{Items: pbItems}, nil
}
