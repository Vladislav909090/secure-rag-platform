package usecase

import (
	"context"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

func (uc *DocumentUsecase) getActiveDocument(ctx context.Context, docUUID string) (*model.Document, error) {
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

	return doc, nil
}
