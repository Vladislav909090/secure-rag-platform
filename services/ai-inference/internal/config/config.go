package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Key string

const (
	GRPCPort         Key = "GRPC_PORT"
	ModelsJSON       Key = "AI_INFERENCE_MODELS_JSON"
	OpenAIAPIKey     Key = "OPENAI_API_KEY"
	DefaultOpenAIURL Key = "AI_INFERENCE_DEFAULT_OPENAI_BASE_URL"
)

type TaskType string

const (
	TaskGeneration TaskType = "generation"
	TaskEmbedding  TaskType = "embedding"
)

type GenerationDefaults struct {
	Temperature      *float32 `json:"temperature,omitempty"`
	TopP             *float32 `json:"top_p,omitempty"`
	MaxTokens        *int32   `json:"max_tokens,omitempty"`
	PresencePenalty  *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32 `json:"frequency_penalty,omitempty"`
}

type ModelAlias struct {
	Task               TaskType           `json:"task"`
	Provider           string             `json:"provider"`
	Model              string             `json:"model"`
	BaseURL            string             `json:"base_url"`
	APIKey             string             `json:"api_key,omitempty"`
	GenerationDefaults GenerationDefaults `json:"generation_defaults,omitempty"`
}

type Runtime struct {
	GRPCPort     string
	ModelAliases map[string]ModelAlias
}

func GetValue(key Key) string {
	return os.Getenv(string(key))
}

func Load() (*Runtime, error) {
	aliases := defaultModelAliases()

	if raw := strings.TrimSpace(GetValue(ModelsJSON)); raw != "" {
		parsed := make(map[string]ModelAlias)
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, fmt.Errorf("parse AI_INFERENCE_MODELS_JSON: %w", err)
		}
		aliases = parsed
	}

	for aliasName, alias := range aliases {
		if strings.TrimSpace(alias.APIKey) == "" {
			alias.APIKey = strings.TrimSpace(GetValue(OpenAIAPIKey))
			aliases[aliasName] = alias
		}
	}

	if err := validateAliases(aliases); err != nil {
		return nil, err
	}

	return &Runtime{
		GRPCPort:     strings.TrimSpace(GetValue(GRPCPort)),
		ModelAliases: aliases,
	}, nil
}

func validateAliases(aliases map[string]ModelAlias) error {
	if len(aliases) == 0 {
		return errors.New("model aliases are empty")
	}

	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		alias := aliases[name]
		if strings.TrimSpace(name) == "" {
			return errors.New("model alias cannot be empty")
		}

		switch alias.Task {
		case TaskGeneration, TaskEmbedding:
		default:
			return fmt.Errorf("alias %q has unsupported task %q", name, alias.Task)
		}

		if strings.TrimSpace(alias.Provider) == "" {
			return fmt.Errorf("alias %q provider is empty", name)
		}
		if strings.TrimSpace(alias.Model) == "" {
			return fmt.Errorf("alias %q model is empty", name)
		}
		if strings.TrimSpace(alias.BaseURL) == "" {
			return fmt.Errorf("alias %q base_url is empty", name)
		}
	}

	return nil
}

func defaultModelAliases() map[string]ModelAlias {
	defaultURL := strings.TrimSpace(GetValue(DefaultOpenAIURL))
	if defaultURL == "" {
		defaultURL = "https://api.openai.com/v1"
	}

	return map[string]ModelAlias{
		"chat.default": {
			Task:     TaskGeneration,
			Provider: "openai_compat",
			Model:    "gpt-4o-mini",
			BaseURL:  defaultURL,
			GenerationDefaults: GenerationDefaults{
				Temperature: ptrFloat32(0.2),
				TopP:        ptrFloat32(1.0),
				MaxTokens:   ptrInt32(1024),
			},
		},
		"chat.quality": {
			Task:     TaskGeneration,
			Provider: "openai_compat",
			Model:    "gpt-4.1",
			BaseURL:  defaultURL,
			GenerationDefaults: GenerationDefaults{
				Temperature: ptrFloat32(0.2),
				TopP:        ptrFloat32(1.0),
				MaxTokens:   ptrInt32(1024),
			},
		},
		"embed.default": {
			Task:     TaskEmbedding,
			Provider: "openai_compat",
			Model:    "text-embedding-3-small",
			BaseURL:  defaultURL,
		},
	}
}

func ptrFloat32(v float32) *float32 {
	return &v
}

func ptrInt32(v int32) *int32 {
	return &v
}
