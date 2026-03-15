package usecase

import (
	"context"
	"fmt"
)

// DownloadVersionFile скачивает файл конкретной версии.
func (uc *DocumentUsecase) DownloadVersionFile(ctx context.Context, docUUID string, versionNumber int32) (*FileDownload, error) {
	doc, err := uc.getActiveDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	ver, err := uc.repo.GetVersion(ctx, doc.ID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}

	if ver == nil {
		return nil, ErrVersionNotFound
	}

	body, err := uc.storage.Download(ctx, ver.StorageKey)
	if err != nil {
		return nil, ErrFileNotFound
	}

	return &FileDownload{
		Body:      body,
		FileName:  ver.FileName,
		MimeType:  ver.MimeType,
		SizeBytes: ver.SizeBytes,
	}, nil
}
