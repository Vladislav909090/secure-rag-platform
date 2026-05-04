package usecase

import (
	"context"
	"fmt"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

// UpdateAttributes полностью заменяет attributes.
func (uc *DocumentUsecase) UpdateAttributes(
	ctx context.Context,
	docUUID string,
	attributes map[string]any,
) (*model.Document, error) {
	if _, err := uc.getActiveDocument(ctx, docUUID); err != nil {
		return nil, err
	}

	if attributes == nil {
		return nil, ErrInvalidRequest
	}

	now := time.Now().UTC()
	if err := uc.repo.UpdateAttributes(ctx, docUUID, attributes, now); err != nil {
		return nil, fmt.Errorf("update attributes: %w", err)
	}

	doc, err := uc.repo.GetDocumentByUUID(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("re-fetch document: %w", err)
	}

	return doc, nil
}
