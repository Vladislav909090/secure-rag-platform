package usecase

import (
	"context"
	"fmt"
	"strings"

	"secure-rag-platform/services/ai-inference/internal/config"
)

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if strings.TrimSpace(req.ModelAlias) == "" {
		return nil, ErrInvalidArgument
	}
	if len(req.Messages) == 0 {
		return nil, ErrInvalidArgument
	}
	for _, msg := range req.Messages {
		if strings.TrimSpace(msg.Role) == "" || strings.TrimSpace(msg.Content) == "" {
			return nil, ErrInvalidArgument
		}
	}

	alias, provider, err := s.resolve(req.ModelAlias, config.TaskGeneration)
	if err != nil {
		return nil, err
	}

	req.Params = mergeGenerationParams(alias.GenerationDefaults, req.Params)

	s.logger.InfoContext(ctx, "запрос generation",
		"component", "ai-inference.generation",
		"request_id", req.RequestID,
		"alias", req.ModelAlias,
		"provider", alias.Provider,
		"model", alias.Model,
	)

	result, err := provider.Generate(ctx, alias, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}

	result.ResolvedProvider = alias.Provider
	result.ResolvedModel = alias.Model

	return result, nil
}
