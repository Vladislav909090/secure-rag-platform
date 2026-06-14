package grpc

import (
	"context"

	pb "secure-rag-platform/api/gen/go/gateway/v1"
	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/genproto/googleapis/api/httpbody"
)

func (s *Server) ListDocuments(
	ctx context.Context,
	req *pb.ListDocumentsRequest,
) (*pb.ListDocumentsResponse, error) {
	_ = req

	if err := s.requireUC(); err != nil {
		return nil, err
	}

	items, err := s.uc.ListDocuments(ctx, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	out := make([]*pb.DocumentItem, 0, len(items))
	for _, item := range items {
		out = append(out, documentItemToProto(item))
	}
	return &pb.ListDocumentsResponse{Items: out}, nil
}

func (s *Server) GetDocument(
	ctx context.Context,
	req *pb.GetDocumentRequest,
) (*pb.GetDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	item, err := s.uc.GetDocument(ctx, req.GetDocumentUuid(), extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetDocumentResponse{
		Document: documentToProto(item.Document),
	}, nil
}

func (s *Server) DownloadFile(
	ctx context.Context,
	req *pb.DownloadFileRequest,
) (*httpbody.HttpBody, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	file, err := s.uc.DownloadFile(ctx, req.GetDocumentUuid(), extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &httpbody.HttpBody{ContentType: file.ContentType, Data: file.Data}, nil
}

func (s *Server) DeleteDocument(
	ctx context.Context,
	req *pb.DeleteDocumentRequest,
) (*pb.DeleteDocumentResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	result, err := s.uc.DeleteDocument(ctx, req.GetDocumentUuid(), extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteDocumentResponse{
		DocumentUuid: result.DocumentUUID,
		Deleted:      result.Deleted,
		DeletedAt:    result.DeletedAt,
		IndexDeleted: result.IndexDeleted,
	}, nil
}

func documentItemToProto(item usecase.DocumentItem) *pb.DocumentItem {
	return &pb.DocumentItem{
		Document: documentToProto(item.Document),
	}
}

func documentToProto(doc usecase.Document) *pb.Document {
	return &pb.Document{
		Id:             doc.ID,
		Uuid:           doc.UUID,
		Title:          doc.Title,
		Description:    doc.Description,
		Attributes:     mapToStruct(doc.Attributes),
		FileName:       doc.FileName,
		FileExtension:  doc.FileExtension,
		MimeType:       doc.MimeType,
		SizeBytes:      doc.SizeBytes,
		ChecksumSha256: doc.ChecksumSHA256,
		StorageKey:     doc.StorageKey,
		IndexStatus:    doc.IndexStatus,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
		DeletedAt:      doc.DeletedAt,
	}
}
