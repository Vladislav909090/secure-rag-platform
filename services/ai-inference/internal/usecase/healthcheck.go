package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"secure-rag-platform/services/ai-inference/internal/config"
)

func (s *Service) CheckDependencies(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("service is not ready: aliases=0 providers=0")
	}
	if !s.Ready() {
		return fmt.Errorf("service is not ready: aliases=%d providers=%d", len(s.aliases), len(s.providers))
	}
	if s.dependencyChecksDisabled {
		return nil
	}

	aliasNames := make([]string, 0, len(s.aliases))
	for aliasName := range s.aliases {
		aliasNames = append(aliasNames, aliasName)
	}
	sort.Strings(aliasNames)

	var hasGeneration bool
	var hasEmbedding bool
	issues := make([]string, 0)

	for _, aliasName := range aliasNames {
		alias := s.aliases[aliasName]

		switch alias.Task {
		case config.TaskGeneration:
			hasGeneration = true
			_, err := s.Generate(ctx, GenerateRequest{
				RequestID:  "health-generate-" + aliasName,
				ModelAlias: aliasName,
				Messages: []Message{
					{Role: "user", Content: "health check"},
				},
				Params: GenerationParams{MaxTokens: int32Ptr(1)},
			})
			if err != nil {
				issues = append(issues, fmt.Sprintf("generate alias %q failed: %v", aliasName, err))
			}
		case config.TaskEmbedding:
			hasEmbedding = true
			_, err := s.Embed(ctx, BatchEmbedRequest{
				RequestID:  "health-embed-" + aliasName,
				ModelAlias: aliasName,
				Texts:      []string{"health check"},
				Normalize:  false,
			})
			if err != nil {
				issues = append(issues, fmt.Sprintf("embed alias %q failed: %v", aliasName, err))
			}
		default:
			issues = append(issues, fmt.Sprintf("alias %q has unsupported task %q", aliasName, alias.Task))
		}
	}

	if !hasGeneration {
		issues = append(issues, "no generation aliases configured")
	}
	if !hasEmbedding {
		issues = append(issues, "no embedding aliases configured")
	}

	if len(issues) > 0 {
		return fmt.Errorf("health check failed: %s", strings.Join(issues, "; "))
	}

	return nil
}

func int32Ptr(v int32) *int32 {
	return &v
}
