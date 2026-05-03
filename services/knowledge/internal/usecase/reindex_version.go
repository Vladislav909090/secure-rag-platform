package usecase

import (
	"context"
	"fmt"

	"secure-rag-platform/services/knowledge/internal/model"
)

// ReindexVersion ставит статус PENDING и инициирует переиндексацию.
func (uc *DocumentUsecase) ReindexVersion(
	ctx context.Context,
	docUUID string,
	versionNumber int32,
) (*ReindexOutput, error) {
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

	if err := uc.repo.UpdateIndexStatus(ctx, ver.ID, model.IndexStatusPending); err != nil {
		return nil, fmt.Errorf("update index status: %w", err)
	}

	return &ReindexOutput{
		DocumentUUID:  docUUID,
		VersionNumber: versionNumber,
		IndexStatus:   model.IndexStatusPending,
	}, nil
}
