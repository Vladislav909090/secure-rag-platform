package grpc

import (
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

func TestKnowledgeDocumentToProto(t *testing.T) {
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
