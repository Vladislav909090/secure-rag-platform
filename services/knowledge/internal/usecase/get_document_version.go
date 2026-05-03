package usecase

import (
	"context"
	"fmt"
)

// GetDocumentVersion возвращает конкретную версию.
func (uc *DocumentUsecase) GetDocumentVersion(
	ctx context.Context,
	docUUID string,
	versionNumber int32,
) (*DocumentVersionDetail, error) {
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

	ver, err := uc.repo.GetVersion(ctx, doc.ID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}

	if ver == nil {
		return nil, ErrVersionNotFound
	}

	return &DocumentVersionDetail{Document: doc, Version: ver}, nil
}
