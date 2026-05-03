package usecase

import (
	"context"
	"fmt"
	"strings"

	ragv1 "secure-rag-platform/services/rag/gen/v1"
)

// IndexDocumentVersion инициирует индексацию документа через RAG.
func (s *Service) IndexDocumentVersion(ctx context.Context, req IndexRequest, accessToken string) (*IndexResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	_, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	docUUID := strings.TrimSpace(req.DocumentUUID)
	if docUUID == "" {
		return nil, ErrInvalidRequest
	}
	if req.VersionNumber <= 0 {
		return nil, ErrInvalidRequest
	}

	resp, err := s.rag.IndexDocumentVersion(ctx, &ragv1.IndexDocumentVersionRequest{
		DocumentUuid:        docUUID,
		VersionNumber:       req.VersionNumber,
		EmbeddingModelAlias: req.EmbeddingModelAlias,
		ChunkSize:           req.ChunkSize,
		ChunkOverlap:        req.ChunkOverlap,
	})
	if err != nil {
		return nil, fmt.Errorf("rag index: %w", err)
	}

	return &IndexResult{
		DocumentUUID:           resp.GetDocumentUuid(),
		VersionNumber:          resp.GetVersionNumber(),
		ChunkCount:             resp.GetChunkCount(),
		EmbeddingDimension:     resp.GetEmbeddingDimension(),
		ResolvedEmbeddingModel: resp.GetResolvedEmbeddingModel(),
	}, nil
}
