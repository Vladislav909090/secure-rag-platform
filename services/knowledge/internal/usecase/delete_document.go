package usecase

import (
	"context"
	"fmt"
	"time"
)

// DeleteDocument мягко удаляет документ
func (uc *DocumentUsecase) DeleteDocument(ctx context.Context, docUUID string) (*DeleteDocumentOutput, error) {
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

	now := time.Now().UTC()
	if err := uc.repo.SoftDeleteDocument(ctx, docUUID, now); err != nil {
		return nil, fmt.Errorf("soft delete: %w", err)
	}

	return &DeleteDocumentOutput{DocumentUUID: docUUID, DeletedAt: now}, nil
}
