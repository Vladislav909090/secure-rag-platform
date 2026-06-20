package usecase

import (
	"context"
	"fmt"
)

// ListDocuments возвращает активные документы
func (uc *DocumentUsecase) ListDocuments(ctx context.Context) ([]*DocumentDetail, error) {
	docs, err := uc.repo.ListActiveDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	result := make([]*DocumentDetail, 0, len(docs))
	for _, doc := range docs {
		result = append(result, &DocumentDetail{Document: doc})
	}

	return result, nil
}
