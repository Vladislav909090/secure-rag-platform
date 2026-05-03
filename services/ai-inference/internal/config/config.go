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
	GRPCPort        Key = "GRPC_PORT"
	ProviderTimeout Key = "AI_INFERENCE_PROVIDER_TIMEOUT"
)

const (
	DefaultModelsConfigPath  = "config/models.json"
	OpenAICompatProviderName = "openai_compat"
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
	GRPCPort        string
	ProviderTimeout string
	ModelAliases    map[string]ModelAlias
}

func GetValue(key Key) string {
	return os.Getenv(string(key))
}

func Load() (*Runtime, error) {
	return LoadFromFile(DefaultModelsConfigPath)
}

func LoadFromFile(modelsConfigPath string) (*Runtime, error) {
	modelsConfigPath = strings.TrimSpace(modelsConfigPath)
	if modelsConfigPath == "" {
		modelsConfigPath = DefaultModelsConfigPath
	}

	raw, err := os.ReadFile(modelsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read models config %q: %w", modelsConfigPath, err)
	}

	aliases := make(map[string]ModelAlias)
	if err := json.Unmarshal(raw, &aliases); err != nil {
		return nil, fmt.Errorf("parse models config %q: %w", modelsConfigPath, err)
	}

	for aliasName, alias := range aliases {
		if strings.TrimSpace(alias.Provider) == "" {
			alias.Provider = OpenAICompatProviderName
			aliases[aliasName] = alias
		}
	}

	if err := validateAliases(aliases); err != nil {
		return nil, err
	}

	return &Runtime{
		GRPCPort:        strings.TrimSpace(GetValue(GRPCPort)),
		ProviderTimeout: strings.TrimSpace(GetValue(ProviderTimeout)),
		ModelAliases:    aliases,
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
		if alias.Provider != OpenAICompatProviderName {
			return fmt.Errorf("alias %q provider %q is not supported (only %q)", name, alias.Provider, OpenAICompatProviderName)
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
