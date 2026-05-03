package grpc

import (
	"context"

	pb "secure-rag-platform/services/gateway/gen/v1"
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

	out := make([]*pb.DocumentWithVersions, 0, len(items))
	for _, item := range items {
		out = append(out, documentWithVersionsToProto(item))
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
		Versions: versionsToProto(item.Versions),
	}, nil
}

func (s *Server) GetDocumentVersion(
	ctx context.Context,
	req *pb.GetDocumentVersionRequest,
) (*pb.GetDocumentVersionResponse, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	item, err := s.uc.GetDocumentVersion(
		ctx,
		req.GetDocumentUuid(),
		req.GetVersionNumber(),
		extractAccessToken(ctx),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var version *pb.DocumentVersion
	if len(item.Versions) > 0 {
		version = versionToProto(item.Versions[0])
	}

	return &pb.GetDocumentVersionResponse{
		Document: documentToProto(item.Document),
		Version:  version,
	}, nil
}

func (s *Server) DownloadFile(
	ctx context.Context,
	req *pb.DownloadFileRequest,
) (*httpbody.HttpBody, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	file, err := s.uc.DownloadFile(ctx, req.GetDocumentUuid(), 0, extractAccessToken(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &httpbody.HttpBody{ContentType: file.ContentType, Data: file.Data}, nil
}

func (s *Server) DownloadVersionFile(
	ctx context.Context,
	req *pb.DownloadVersionFileRequest,
) (*httpbody.HttpBody, error) {
	if err := s.requireUC(); err != nil {
		return nil, err
	}

	file, err := s.uc.DownloadFile(
		ctx,
		req.GetDocumentUuid(),
		req.GetVersionNumber(),
		extractAccessToken(ctx),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &httpbody.HttpBody{ContentType: file.ContentType, Data: file.Data}, nil
}

func documentWithVersionsToProto(item usecase.DocumentWithVersions) *pb.DocumentWithVersions {
	return &pb.DocumentWithVersions{
		Document: documentToProto(item.Document),
		Versions: versionsToProto(item.Versions),
	}
}

func documentToProto(doc usecase.Document) *pb.Document {
	return &pb.Document{
		Id:                   doc.ID,
		Uuid:                 doc.UUID,
		Title:                doc.Title,
		Description:          doc.Description,
		Attributes:           mapToStruct(doc.Attributes),
		CurrentVersionNumber: doc.CurrentVersionNumber,
		CreatedAt:            doc.CreatedAt,
		UpdatedAt:            doc.UpdatedAt,
		DeletedAt:            doc.DeletedAt,
	}
}

func versionsToProto(versions []usecase.DocumentVersion) []*pb.DocumentVersion {
	out := make([]*pb.DocumentVersion, 0, len(versions))
	for _, version := range versions {
		out = append(out, versionToProto(version))
	}
	return out
}

func versionToProto(version usecase.DocumentVersion) *pb.DocumentVersion {
	return &pb.DocumentVersion{
		Id:             version.ID,
		Uuid:           version.UUID,
		DocumentId:     version.DocumentID,
		VersionNumber:  version.VersionNumber,
		FileName:       version.FileName,
		FileExtension:  version.FileExtension,
		MimeType:       version.MimeType,
		SizeBytes:      version.SizeBytes,
		ChecksumSha256: version.ChecksumSHA256,
		StorageKey:     version.StorageKey,
		IndexStatus:    version.IndexStatus,
		CreatedAt:      version.CreatedAt,
	}
}
