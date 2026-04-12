package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) ListDocuments(ctx context.Context, _ *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	items, err := s.uc.ListDocuments(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbItems := make([]*pb.DocumentWithVersions, 0, len(items))
	for _, item := range items {
		pbVersions := make([]*pb.DocumentVersion, 0, len(item.Versions))
		for _, v := range item.Versions {
			pbVersions = append(pbVersions, versionToProto(v))
		}
		pbItems = append(pbItems, &pb.DocumentWithVersions{
			Document: documentToProto(item.Document),
			Versions: pbVersions,
		})
	}

	return &pb.ListDocumentsResponse{Items: pbItems}, nil
}
