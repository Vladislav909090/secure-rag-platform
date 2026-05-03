package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) ReindexVersion(
	ctx context.Context,
	req *pb.ReindexVersionRequest,
) (*pb.ReindexVersionResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	result, err := s.uc.ReindexVersion(ctx, req.GetDocumentUuid(), req.GetVersionNumber())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ReindexVersionResponse{
		DocumentUuid:  result.DocumentUUID,
		VersionNumber: result.VersionNumber,
		IndexStatus:   result.IndexStatus,
	}, nil
}
