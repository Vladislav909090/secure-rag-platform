package usecase

import (
	"context"
)

// DownloadFile скачивает файл документа
func (uc *DocumentUsecase) DownloadFile(ctx context.Context, docUUID string) (*FileDownload, error) {
	doc, err := uc.getActiveDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	body, err := uc.storage.Download(ctx, doc.StorageKey)
	if err != nil {
		return nil, ErrFileNotFound
	}

	return &FileDownload{
		Body:      body,
		FileName:  doc.FileName,
		MimeType:  doc.MimeType,
		SizeBytes: doc.SizeBytes,
	}, nil
}
