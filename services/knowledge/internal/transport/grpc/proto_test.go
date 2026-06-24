package grpc

import (
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeDocumentToProto(t *testing.T) {
	t.Parallel()

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

	require.NotNil(t, doc)
	assert.Equal(t, "doc-1", doc.GetUuid())
	assert.Equal(t, desc, doc.GetDescription())
	assert.Equal(t, "public", doc.GetAttributes().AsMap()["level"])
	assert.Equal(t, deletedAt.Format(time.RFC3339), doc.GetDeletedAt())
}
