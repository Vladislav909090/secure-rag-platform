package usecase

import (
	"context"
	"fmt"
	"strings"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	ragv1 "secure-rag-platform/api/gen/go/rag/v1"
)

// ReindexDocument переиндексирует один документ через RAG. Доступно редактору документов.
func (s *Service) ReindexDocument(
	ctx context.Context,
	req ReindexRequest,
	accessToken string,
) (*ReindexResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	docUUID := strings.TrimSpace(req.DocumentUUID)
	if docUUID == "" {
		return nil, ErrInvalidRequest
	}
	doc, err := s.getAllowedDocument(ctx, docUUID, subject)
	if err != nil {
		return nil, err
	}
	docUUID = doc.GetUuid()

	_, err = s.knowledge.ReindexDocument(ctx, &knowledgev1.ReindexDocumentRequest{
		DocumentUuid: docUUID,
	})
	if err != nil {
		return nil, mapUpstreamError(err, "reindex document")
	}

	resp, err := s.rag.IndexDocument(ctx, &ragv1.IndexDocumentRequest{
		DocumentUuid:        docUUID,
		EmbeddingModelAlias: req.EmbeddingModelAlias,
		ChunkSize:           req.ChunkSize,
		ChunkOverlap:        req.ChunkOverlap,
	})
	if err != nil {
		return nil, fmt.Errorf("rag index: %w", err)
	}

	return &ReindexResult{
		DocumentUUID:           resp.GetDocumentUuid(),
		ChunkCount:             resp.GetChunkCount(),
		EmbeddingDimension:     resp.GetEmbeddingDimension(),
		ResolvedEmbeddingModel: resp.GetResolvedEmbeddingModel(),
	}, nil
}

// ReindexAllDocuments переиндексирует доступные активные документы. Доступно редактору документов.
func (s *Service) ReindexAllDocuments(
	ctx context.Context,
	req ReindexRequest,
	accessToken string,
) (*ReindexAllResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	listResp, err := s.knowledge.ListDocuments(ctx, &knowledgev1.ListDocumentsRequest{})
	if err != nil {
		return nil, mapUpstreamError(err, "list documents")
	}

	result := &ReindexAllResult{
		Items: make([]ReindexItemResult, 0, len(listResp.GetItems())),
	}

	for _, item := range listResp.GetItems() {
		doc := item.GetDocument()
		if doc == nil {
			continue
		}
		allowed, err := s.isDocumentAllowed(ctx, subject, doc)
		if err != nil {
			return nil, err
		}
		if !allowed {
			continue
		}

		result.TotalCount++
		docUUID := doc.GetUuid()
		itemResult := ReindexItemResult{DocumentUUID: docUUID}

		_, err = s.knowledge.ReindexDocument(ctx, &knowledgev1.ReindexDocumentRequest{DocumentUuid: docUUID})
		if err != nil {
			itemResult.Error = mapUpstreamError(err, "reindex document").Error()
			result.FailedCount++
			result.Items = append(result.Items, itemResult)
			continue
		}

		var indexResp *ragv1.IndexDocumentResponse
		indexResp, err = s.rag.IndexDocument(ctx, &ragv1.IndexDocumentRequest{
			DocumentUuid:        docUUID,
			EmbeddingModelAlias: req.EmbeddingModelAlias,
			ChunkSize:           req.ChunkSize,
			ChunkOverlap:        req.ChunkOverlap,
		})
		if err != nil {
			itemResult.Error = fmt.Errorf("rag index: %w", err).Error()
			result.FailedCount++
			result.Items = append(result.Items, itemResult)
			continue
		}

		itemResult.Indexed = true
		itemResult.ChunkCount = indexResp.GetChunkCount()
		itemResult.EmbeddingDimension = indexResp.GetEmbeddingDimension()
		itemResult.ResolvedEmbeddingModel = indexResp.GetResolvedEmbeddingModel()
		result.IndexedCount++
		result.Items = append(result.Items, itemResult)
	}

	return result, nil
}
