package grpc

import (
	"context"

	pb "secure-rag-platform/services/knowledge/gen/v1"
)

func (s *KnowledgeServiceServerImpl) GetDocumentVersion(ctx context.Context, req *pb.GetDocumentVersionRequest) (*pb.GetDocumentVersionResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}
	result, err := s.uc.GetDocumentVersion(ctx, req.GetDocumentUuid(), req.GetVersionNumber())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetDocumentVersionResponse{
		Document: documentToProto(result.Document),
		Version:  versionToProto(result.Version),
	}, nil
}
