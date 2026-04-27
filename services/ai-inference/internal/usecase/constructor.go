package usecase

import (
	"log"

	"secure-rag-platform/services/ai-inference/internal/config"
)

type Service struct {
	aliases   map[string]config.ModelAlias
	providers map[string]Provider
	logger    *log.Logger
}

func NewService(aliases map[string]config.ModelAlias, providers []Provider, logger *log.Logger) *Service {
	aliasCopy := make(map[string]config.ModelAlias, len(aliases))
	for name, alias := range aliases {
		aliasCopy[name] = alias
	}

	providerMap := make(map[string]Provider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		providerMap[provider.Name()] = provider
	}

	if logger == nil {
		logger = log.Default()
	}

	return &Service{
		aliases:   aliasCopy,
		providers: providerMap,
		logger:    logger,
	}
}

func (s *Service) Ready() bool {
	if s == nil {
		return false
	}

	return len(s.aliases) > 0 && len(s.providers) > 0
}

func (s *Service) resolve(aliasName string, expectedTask config.TaskType) (config.ModelAlias, Provider, error) {
	alias, ok := s.aliases[aliasName]
	if !ok {
		return config.ModelAlias{}, nil, ErrAliasNotFound
	}
	if alias.Task != expectedTask {
		return config.ModelAlias{}, nil, ErrAliasTaskMismatch
	}

	provider, ok := s.providers[alias.Provider]
	if !ok {
		return config.ModelAlias{}, nil, ErrProviderNotConfigured
	}

	return alias, provider, nil
}
