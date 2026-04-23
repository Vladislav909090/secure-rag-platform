package usecase

import (
	"sort"

	"secure-rag-platform/services/ai-inference/internal/config"
)

func (s *Service) ListModels(task config.TaskType) []ModelInfo {
	out := make([]ModelInfo, 0, len(s.aliases))
	for aliasName, alias := range s.aliases {
		if task != "" && alias.Task != task {
			continue
		}
		out = append(out, ModelInfo{
			Alias:    aliasName,
			Task:     alias.Task,
			Provider: alias.Provider,
			Model:    alias.Model,
			BaseURL:  alias.BaseURL,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Alias < out[j].Alias
	})

	return out
}
