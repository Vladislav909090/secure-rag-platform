package grpc

import (
	"errors"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDocumentToProto(t *testing.T) {
	desc := "description"
	deletedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	doc := documentToProto(&model.Document{
		ID:             1,
		UUID:           "doc-1",
		Title:          "title",
		Description:    &desc,
		Attributes:     map[string]any{"level": "public"},
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      10,
		ChecksumSHA256: "abc",
		StorageKey:     "documents/doc-1/file",
		IndexStatus:    model.IndexStatusPending,
		CreatedAt:      deletedAt,
		UpdatedAt:      deletedAt,
		DeletedAt:      &deletedAt,
	})

	if doc.GetUuid() != "doc-1" || doc.GetDescription() != desc || doc.GetAttributes().AsMap()["level"] != "public" {
		t.Fatalf("unexpected proto document: %#v", doc)
	}
	if doc.GetDeletedAt() != deletedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected deleted_at %q", doc.GetDeletedAt())
	}
}

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrDocumentNotFound, codes.NotFound},
		{usecase.ErrFileNotFound, codes.NotFound},
		{usecase.ErrDocumentDeleted, codes.Aborted},
		{usecase.ErrDocumentNotDeleted, codes.Aborted},
		{usecase.ErrFileTooLarge, codes.ResourceExhausted},
		{usecase.ErrInvalidRequest, codes.InvalidArgument},
		{errors.New("upload to storage: down"), codes.Internal},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
