package usecase

import (
	"context"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

// ReindexDocument ставит статус PENDING перед переиндексацией
func (uc *DocumentUsecase) ReindexDocument(
	ctx context.Context,
	docUUID string,
) (*ReindexOutput, error) {
	doc, err := uc.getActiveDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateIndexStatus(ctx, doc.ID, model.IndexStatusPending); err != nil {
		return nil, fmt.Errorf("update index status: %w", err)
	}

	return &ReindexOutput{
		DocumentUUID: docUUID,
		IndexStatus:  model.IndexStatusPending,
	}, nil
}
