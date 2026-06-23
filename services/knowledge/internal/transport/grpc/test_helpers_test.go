package grpc

import (
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

func knowledgeTestDocument(uuid string) *model.Document {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	desc := "description"

	return &model.Document{
		ID:             1,
		UUID:           uuid,
		Title:          "title",
		Description:    &desc,
		Attributes:     map[string]any{"department": "search"},
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      9,
		ChecksumSHA256: "checksum",
		StorageKey:     "documents/" + uuid + "/file",
		IndexStatus:    model.IndexStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
