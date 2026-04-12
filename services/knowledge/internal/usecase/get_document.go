package usecase

import (
	"context"
	"fmt"
)

// GetDocument возвращает документ с версиями.
func (uc *DocumentUsecase) GetDocument(ctx context.Context, docUUID string) (*DocumentWithVersions, error) {
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

	versions, err := uc.repo.GetVersionsByDocumentID(ctx, doc.ID)
	if err != nil {
		return nil, fmt.Errorf("get versions: %w", err)
	}

	return &DocumentWithVersions{Document: doc, Versions: versions}, nil
}
