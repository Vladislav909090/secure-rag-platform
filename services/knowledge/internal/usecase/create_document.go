package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/google/uuid"
)

// CreateDocument создаёт документ с единственным файлом
func (uc *DocumentUsecase) CreateDocument(
	ctx context.Context,
	input CreateDocumentInput,
	file io.Reader,
	fileName string,
	mimeType string,
) (*CreateDocumentOutput, error) {
	if input.Title == "" {
		return nil, ErrInvalidRequest
	}

	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	now := time.Now().UTC()
	docUUID := uuid.New().String()
	storageKey := fmt.Sprintf("documents/%s/file", docUUID)

	sizeBytes, checksum, err := uc.uploadAndHash(ctx, storageKey, file, mimeType)
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	attrs := input.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}

	doc := &model.Document{
		UUID:           docUUID,
		Title:          input.Title,
		Description:    input.Description,
		Attributes:     attrs,
		FileName:       fileName,
		FileExtension:  ext,
		MimeType:       mimeType,
		SizeBytes:      sizeBytes,
		ChecksumSHA256: checksum,
		StorageKey:     storageKey,
		IndexStatus:    model.IndexStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.repo.CreateDocument(ctx, doc); err != nil {
		_ = uc.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("create document: %w", err)
	}

	return &CreateDocumentOutput{Document: doc}, nil
}
