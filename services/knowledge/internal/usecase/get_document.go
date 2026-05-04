package usecase

import (
	"context"
	"fmt"
)

// GetDocument возвращает документ.
func (uc *DocumentUsecase) GetDocument(ctx context.Context, docUUID string) (*DocumentDetail, error) {
	doc, err := uc.repo.GetDocumentByUUID(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	if doc == nil {
		return nil, ErrDocumentNotFound
	}

	if doc.DeletedAt != nil {
		return nil, ErrDocumentDeleted
	}

	return &DocumentDetail{Document: doc}, nil
}
