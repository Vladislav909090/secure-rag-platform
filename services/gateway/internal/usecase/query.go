package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	ragv1 "secure-rag-platform/api/gen/go/rag/v1"
)

// Query выполняет поиск и генерацию ответа через RAG
func (s *Service) Query(ctx context.Context, req QueryRequest, accessToken string) (*QueryResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	question := strings.TrimSpace(req.Query)
	if question == "" {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	documentUUIDs, err := s.resolveDocuments(ctx, req.DocumentUUIDs, subject)
	if err != nil {
		return nil, err
	}
	if !s.disableFilter && len(documentUUIDs) == 0 {
		return &QueryResult{
			Answer: "No documents available for this user.",
		}, nil
	}

	topK := req.TopK
	if topK <= 0 {
		topK = s.defaults.TopK
	}

	embedAlias := strings.TrimSpace(req.EmbeddingModelAlias)
	if embedAlias == "" {
		embedAlias = s.defaults.EmbeddingModelAlias
	}

	genAlias := strings.TrimSpace(req.GenerationModelAlias)
	if genAlias == "" {
		genAlias = s.defaults.GenerationModelAlias
	}

	resp, err := s.rag.Query(ctx, &ragv1.QueryRequest{
		Query:                question,
		TopK:                 topK,
		DocumentUuids:        documentUUIDs,
		EmbeddingModelAlias:  embedAlias,
		GenerationModelAlias: genAlias,
	})
	if err != nil {
		return nil, fmt.Errorf("rag query: %w", err)
	}

	contexts := make([]QueryContext, 0, len(resp.GetContexts()))
	for _, ctxItem := range resp.GetContexts() {
		contexts = append(contexts, QueryContext{
			DocumentUUID: ctxItem.GetDocumentUuid(),
			ChunkIndex:   ctxItem.GetChunkIndex(),
			Text:         ctxItem.GetText(),
			Score:        ctxItem.GetScore(),
		})
	}

	return &QueryResult{
		Answer:                  resp.GetAnswer(),
		Contexts:                contexts,
		ResolvedEmbeddingModel:  resp.GetResolvedEmbeddingModel(),
		ResolvedGenerationModel: resp.GetResolvedGenerationModel(),
	}, nil
}

func (s *Service) resolveDocuments(
	ctx context.Context,
	requested []string,
	subject *iamv1.SubjectContext,
) ([]string, error) {
	if s.disableFilter {
		normalized := make([]string, 0, len(requested))
		for _, id := range requested {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			normalized = append(normalized, id)
		}
		if len(normalized) > 0 {
			return normalized, nil
		}

		if s.knowledge == nil {
			return nil, ErrNotConfigured
		}

		resp, err := s.knowledge.ListDocuments(ctx, &knowledgev1.ListDocumentsRequest{})
		if err != nil {
			return nil, fmt.Errorf("list documents: %w", err)
		}

		all := make([]string, 0, len(resp.GetItems()))
		for _, item := range resp.GetItems() {
			doc := item.GetDocument()
			if doc == nil {
				continue
			}
			id := strings.TrimSpace(doc.GetUuid())
			if id == "" {
				continue
			}
			all = append(all, id)
		}

		return all, nil
	}

	if s.knowledge == nil {
		return nil, ErrNotConfigured
	}

	resp, err := s.knowledge.ListDocuments(ctx, &knowledgev1.ListDocumentsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	requestedSet := make(map[string]struct{}, len(requested))
	for _, id := range requested {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		requestedSet[id] = struct{}{}
	}

	allowed := make([]string, 0)
	for _, item := range resp.GetItems() {
		doc := item.GetDocument()
		if doc == nil {
			continue
		}

		if len(requestedSet) > 0 {
			if _, ok := requestedSet[doc.GetUuid()]; !ok {
				continue
			}
		}

		attrs := map[string]any{}
		if doc.GetAttributes() != nil {
			attrs = doc.GetAttributes().AsMap()
		}

		isAllowed, err := s.allowDocument(ctx, subject, attrs)
		if err != nil {
			return nil, err
		}
		if isAllowed {
			allowed = append(allowed, doc.GetUuid())
		}
	}

	return allowed, nil
}

func (s *Service) allowDocument(
	ctx context.Context,
	subject *iamv1.SubjectContext,
	attrs map[string]any,
) (bool, error) {
	if s.policy == nil {
		return false, ErrPolicyRequired
	}
	allowed, err := s.policy.AllowDocument(ctx, subject, attrs)
	if err != nil {
		if errors.Is(err, ErrPolicyRequired) {
			return false, err
		}

		return false, fmt.Errorf("%w: %v", ErrPolicyUnavailable, err)
	}

	return allowed, nil
}
