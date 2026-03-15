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

// UploadVersion загружает новую версию существующего документа.
func (uc *DocumentUsecase) UploadVersion(ctx context.Context, docUUID string, file io.Reader, fileName string, mimeType string) (*UploadVersionOutput, error) {
	doc, err := uc.getActiveDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	now := time.Now().UTC()
	newVersion := doc.CurrentVersionNumber + 1
	verUUID := uuid.New().String()
	storageKey := fmt.Sprintf("documents/%s/v%d", docUUID, newVersion)

	sizeBytes, checksum, err := uc.uploadAndHash(ctx, storageKey, file, mimeType)
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	ver := &model.DocumentVersion{
		UUID:           verUUID,
		DocumentID:     doc.ID,
		VersionNumber:  newVersion,
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

	if err := uc.repo.IncrementVersion(ctx, doc.ID, newVersion, now); err != nil {
		_ = uc.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("increment version: %w", err)
	}

	return &UploadVersionOutput{
		DocumentID:     doc.ID,
		DocumentUUID:   doc.UUID,
		CurrentVersion: newVersion,
		Version:        ver,
	}, nil
}
