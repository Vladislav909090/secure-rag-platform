package usecase

import (
	"context"
	"fmt"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

// RestoreDocument восстанавливает удалённый документ
func (uc *DocumentUsecase) RestoreDocument(ctx context.Context, docUUID string) (*model.Document, error) {
	doc, err := uc.repo.GetDocumentByUUID(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	if doc == nil {
		return nil, ErrDocumentNotFound
	}

	if doc.DeletedAt == nil {
		return nil, ErrDocumentNotDeleted
	}

	now := time.Now().UTC()
	err = uc.repo.RestoreDocument(ctx, docUUID, now)
	if err != nil {
		return nil, fmt.Errorf("restore: %w", err)
	}

	doc, err = uc.repo.GetDocumentByUUID(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("re-fetch document: %w", err)
	}

	return doc, nil
}
