package usecase

import (
	"context"
	"fmt"
	"strings"

	"secure-rag-platform/services/ai-inference/internal/config"
)

func (s *Service) Embed(ctx context.Context, req BatchEmbedRequest) (*BatchEmbedResult, error) {
	if strings.TrimSpace(req.ModelAlias) == "" {
		return nil, ErrInvalidArgument
	}
	if len(req.Texts) == 0 {
		return nil, ErrInvalidArgument
	}
	for _, text := range req.Texts {
		if strings.TrimSpace(text) == "" {
			return nil, ErrInvalidArgument
		}
	}

	alias, provider, err := s.resolve(req.ModelAlias, config.TaskEmbedding)
	if err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "запрос embedding",
		"component", "ai-inference.embedding",
		"request_id", req.RequestID,
		"alias", req.ModelAlias,
		"provider", alias.Provider,
		"model", alias.Model,
		"texts", len(req.Texts),
	)

	result, err := provider.Embed(ctx, alias, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}

	if req.Normalize {
		normalizeVectors(result.Vectors)
	}

	result.ResolvedProvider = alias.Provider
	result.ResolvedModel = alias.Model
	return result, nil
}
