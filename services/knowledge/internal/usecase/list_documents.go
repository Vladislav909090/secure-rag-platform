package usecase

import (
	"context"
	"fmt"
)

// ListDocuments возвращает активные документы с версиями.
func (uc *DocumentUsecase) ListDocuments(ctx context.Context) ([]*DocumentWithVersions, error) {
	docs, err := uc.repo.ListActiveDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	result := make([]*DocumentWithVersions, 0, len(docs))
	for _, doc := range docs {
		versions, err := uc.repo.GetVersionsByDocumentID(ctx, doc.ID)
		if err != nil {
			return nil, fmt.Errorf("get versions for doc %d: %w", doc.ID, err)
		}

		result = append(result, &DocumentWithVersions{Document: doc, Versions: versions})
	}

	return result, nil
}
