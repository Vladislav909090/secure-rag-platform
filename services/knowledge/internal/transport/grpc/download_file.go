package grpc

import (
	"context"
	"fmt"
	"io"

	pb "secure-rag-platform/services/knowledge/gen/v1"

	"google.golang.org/genproto/googleapis/api/httpbody"
)

func (s *KnowledgeServiceServerImpl) DownloadFile(
	ctx context.Context,
	req *pb.DownloadFileRequest,
) (*httpbody.HttpBody, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	dl, err := s.uc.DownloadFile(ctx, req.GetDocumentUuid())
	if err != nil {
		return nil, toGRPCError(err)
	}
	defer dl.Body.Close()

	data, err := io.ReadAll(dl.Body)
	if err != nil {
		return nil, toGRPCError(fmt.Errorf("read file: %w", err))
	}

	return &httpbody.HttpBody{
		ContentType: dl.MimeType,
		Data:        data,
	}, nil
}
