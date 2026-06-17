package usecase

import (
	"context"
	"fmt"
	"strings"
)

// DeleteDocumentIndex удаляет все чанки документа из векторного хранилища
func (s *Service) DeleteDocumentIndex(ctx context.Context, documentUUID string) error {
	if !s.Ready() {
		return ErrNotConfigured
	}

	documentUUID = strings.TrimSpace(documentUUID)
	if documentUUID == "" {
		return ErrInvalidRequest
	}

	if err := s.repo.DeleteChunks(ctx, documentUUID); err != nil {
		return fmt.Errorf("delete chunks: %w", err)
	}

	s.logger.InfoContext(ctx, "индекс документа удалён",
		"component", "rag.index",
		"document_uuid", documentUUID,
	)

	return nil
}
