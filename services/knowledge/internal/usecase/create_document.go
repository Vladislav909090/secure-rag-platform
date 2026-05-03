package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateDocument создаёт документ сразу с первой версией и файлом.
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
	verUUID := uuid.New().String()
	storageKey := fmt.Sprintf("documents/%s/v%d", docUUID, 1)

	sizeBytes, checksum, err := uc.uploadAndHash(ctx, storageKey, file, mimeType)
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	attrs := input.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}

	doc := &model.Document{
		UUID:                 docUUID,
		Title:                input.Title,
		Description:          input.Description,
		Attributes:           attrs,
		CurrentVersionNumber: 1,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := uc.repo.CreateDocument(ctx, doc); err != nil {
		_ = uc.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("create document: %w", err)
	}

	ver := &model.DocumentVersion{
		UUID:           verUUID,
		DocumentID:     doc.ID,
		VersionNumber:  1,
		FileName:       fileName,
		FileExtension:  ext,
		MimeType:       mimeType,
		SizeBytes:      sizeBytes,
		ChecksumSHA256: checksum,
		StorageKey:     storageKey,
		IndexStatus:    model.IndexStatusPending,
		CreatedAt:      now,
	}

	if err := uc.repo.CreateVersion(ctx, ver); err != nil {
		_ = uc.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("create version: %w", err)
	}

	return &CreateDocumentOutput{Document: doc, Version: ver}, nil
}
