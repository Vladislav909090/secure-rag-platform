package grpc

import (
	"testing"

	"secure-rag-platform/services/gateway/internal/usecase"
)

func TestKnowledgeProxyProtoConversions(t *testing.T) {
	doc := documentToProto(usecase.Document{
		ID:             1,
		UUID:           "doc-1",
		Title:          "title",
		Description:    "desc",
		Attributes:     map[string]any{"visibility": "public"},
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      12,
		ChecksumSHA256: "sum",
		StorageKey:     "key",
		IndexStatus:    "READY",
		CreatedAt:      "created",
		UpdatedAt:      "updated",
		DeletedAt:      "deleted",
	})
	if doc.GetUuid() != "doc-1" || doc.GetAttributes().AsMap()["visibility"] != "public" {
		t.Fatalf("unexpected document: %#v", doc)
	}

	item := documentItemToProto(usecase.DocumentItem{Document: usecase.Document{UUID: "doc-2"}})
	if item.GetDocument().GetUuid() != "doc-2" {
		t.Fatalf("unexpected document item: %#v", item)
	}
}
