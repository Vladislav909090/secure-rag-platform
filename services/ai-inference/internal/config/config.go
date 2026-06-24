package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Key string

const (
	HTTPPort        Key = "HTTP_PORT"
	GRPCPort        Key = "GRPC_PORT"
	ProviderTimeout Key = "AI_INFERENCE_PROVIDER_TIMEOUT"

	SkipProviderHealthcheck Key = "AI_INFERENCE_SKIP_PROVIDER_HEALTHCHECK"
	MockProviderResponses   Key = "AI_INFERENCE_MOCK_RESPONSES"
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
	HTTPPort        string
	GRPCPort        string
	ProviderTimeout string
	ModelAliases    map[string]ModelAlias

	SkipProviderHealthcheck bool
	MockProviderResponses   bool
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
	err = json.Unmarshal(raw, &aliases)
	if err != nil {
		return nil, fmt.Errorf("parse models config %q: %w", modelsConfigPath, err)
	}

	for aliasName, alias := range aliases {
		if strings.TrimSpace(alias.Provider) == "" {
			alias.Provider = OpenAICompatProviderName
			aliases[aliasName] = alias
		}
	}

	err = validateAliases(aliases)
	if err != nil {
		return nil, err
	}

	skipProviderHealthcheck, err := parseBoolValue(GetValue(SkipProviderHealthcheck), SkipProviderHealthcheck)
	if err != nil {
		return nil, err
	}
	mockProviderResponses, err := parseBoolValue(GetValue(MockProviderResponses), MockProviderResponses)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		HTTPPort:                strings.TrimSpace(GetValue(HTTPPort)),
		GRPCPort:                strings.TrimSpace(GetValue(GRPCPort)),
		ProviderTimeout:         strings.TrimSpace(GetValue(ProviderTimeout)),
		ModelAliases:            aliases,
		SkipProviderHealthcheck: skipProviderHealthcheck,
		MockProviderResponses:   mockProviderResponses,
	}, nil
}

func parseBoolValue(raw string, key Key) (bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}

	return value, nil
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
