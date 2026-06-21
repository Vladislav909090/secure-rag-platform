package usecase

import (
	"errors"
	"testing"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestDocumentProtoConversionsAndErrorMapping(t *testing.T) {
	attrs, err := structpb.NewStruct(map[string]any{"visibility": "internal"})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	doc := documentFromProto(&knowledgev1.Document{
		Id:             10,
		Uuid:           "doc-1",
		Title:          "policy",
		Description:    "desc",
		Attributes:     attrs,
		FileName:       "policy.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      42,
		ChecksumSha256: "sum",
		StorageKey:     "documents/doc-1/file",
		IndexStatus:    "READY",
		CreatedAt:      "created",
		UpdatedAt:      "updated",
		DeletedAt:      "deleted",
	})
	if doc.UUID != "doc-1" || doc.Attributes["visibility"] != "internal" || doc.SizeBytes != 42 {
		t.Fatalf("unexpected document: %#v", doc)
	}

	item := documentItemFromProto(&knowledgev1.DocumentItem{Document: &knowledgev1.Document{Uuid: "doc-2"}})
	if item.Document.UUID != "doc-2" {
		t.Fatalf("unexpected document item: %#v", item)
	}

	tests := []struct {
		err  error
		want error
	}{
		{status.Error(codes.NotFound, "missing"), ErrNotFound},
		{status.Error(codes.InvalidArgument, "bad"), ErrInvalidRequest},
		{status.Error(codes.Unauthenticated, "auth"), ErrUnauthorized},
		{status.Error(codes.PermissionDenied, "denied"), ErrForbidden},
	}
	for _, tt := range tests {
		if got := mapUpstreamError(tt.err, "operation"); !errors.Is(got, tt.want) {
			t.Fatalf("mapUpstreamError(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}

	unknown := errors.New("upstream exploded")
	if got := mapUpstreamError(unknown, "operation"); got == nil || !errors.Is(got, unknown) {
		t.Fatalf("expected wrapped unknown error, got %v", got)
	}
}
