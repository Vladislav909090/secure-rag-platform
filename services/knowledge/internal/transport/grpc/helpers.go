package grpc

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	pb "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *KnowledgeServiceServerImpl) requireUC() error {
	if s.uc == nil {
		return status.Error(codes.Unavailable, "service not configured")
	}
	return nil
}

func documentToProto(doc *model.Document) *pb.Document {
	d := &pb.Document{
		Id:             doc.ID,
		Uuid:           doc.UUID,
		Title:          doc.Title,
		FileName:       doc.FileName,
		FileExtension:  doc.FileExtension,
		MimeType:       doc.MimeType,
		SizeBytes:      doc.SizeBytes,
		ChecksumSha256: doc.ChecksumSHA256,
		StorageKey:     doc.StorageKey,
		IndexStatus:    doc.IndexStatus,
		CreatedAt:      doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      doc.UpdatedAt.Format(time.RFC3339),
	}
	if doc.Description != nil {
		d.Description = *doc.Description
	}
	if doc.DeletedAt != nil {
		d.DeletedAt = doc.DeletedAt.Format(time.RFC3339)
	}
	if doc.Attributes != nil {
		d.Attributes, _ = structpb.NewStruct(doc.Attributes)
	}
	return d
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrDocumentNotFound),
		errors.Is(err, usecase.ErrFileNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecase.ErrDocumentDeleted),
		errors.Is(err, usecase.ErrDocumentNotDeleted):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, usecase.ErrFileTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(err.Error(), "upload to storage"):
		return status.Error(codes.Internal, err.Error())
	default:
		slog.Error("внутренняя ошибка gRPC",
			"component", "knowledge.grpc",
			"error", err,
		)
		return status.Error(codes.Internal, "internal error")
	}
}
