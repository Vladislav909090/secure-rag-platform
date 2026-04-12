package usecase

import (
	"context"
	"fmt"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

// UpdateDocument обновляет title/description.
func (uc *DocumentUsecase) UpdateDocument(ctx context.Context, docUUID string, title *string, description *string) (*model.Document, error) {
	if _, err := uc.getActiveDocument(ctx, docUUID); err != nil {
		return nil, err
	}

	if title == nil && description == nil {
		return nil, ErrInvalidRequest
	}

	now := time.Now().UTC()
	if err := uc.repo.UpdateDocument(ctx, docUUID, title, description, now); err != nil {
		return nil, fmt.Errorf("update document: %w", err)
	}

	doc, err := uc.repo.GetDocumentByUUID(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("re-fetch document: %w", err)
	}

	return doc, nil
}
