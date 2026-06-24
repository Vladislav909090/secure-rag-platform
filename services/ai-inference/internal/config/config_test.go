package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromFileDefaultsProviderAndFlags(t *testing.T) {
	t.Setenv(string(HTTPPort), "18084")
	t.Setenv(string(GRPCPort), "19094")
	t.Setenv(string(ProviderTimeout), "5s")
	t.Setenv(string(SkipProviderHealthcheck), "true")
	t.Setenv(string(MockProviderResponses), "1")

	path := writeModelsConfig(t, `{
		"chat.default": {
			"task": "generation",
			"model": "chat-model",
			"base_url": "http://provider/v1"
		},
		"embed.default": {
			"task": "embedding",
			"provider": "openai_compat",
			"model": "embed-model",
			"base_url": "http://provider/v1"
		}
	}`)

	runtimeCfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if runtimeCfg.HTTPPort != "18084" || runtimeCfg.GRPCPort != "19094" {
		t.Fatalf("unexpected ports: http=%q grpc=%q", runtimeCfg.HTTPPort, runtimeCfg.GRPCPort)
	}
	if runtimeCfg.ProviderTimeout != "5s" {
		t.Fatalf("unexpected provider timeout %q", runtimeCfg.ProviderTimeout)
	}
	if !runtimeCfg.SkipProviderHealthcheck || !runtimeCfg.MockProviderResponses {
		t.Fatalf("expected healthcheck skip and mock flags to be true")
	}
	if got := runtimeCfg.ModelAliases["chat.default"].Provider; got != OpenAICompatProviderName {
		t.Fatalf("expected default provider %q, got %q", OpenAICompatProviderName, got)
	}
}

func TestLoadFromFileRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "empty aliases",
			raw:  `{}`,
			want: "model aliases are empty",
		},
		{
			name: "unsupported task",
			raw:  `{"bad":{"task":"ranking","provider":"openai_compat","model":"m","base_url":"http://x"}}`,
			want: "unsupported task",
		},
		{
			name: "unsupported provider",
			raw:  `{"bad":{"task":"generation","provider":"other","model":"m","base_url":"http://x"}}`,
			want: "not supported",
		},
		{
			name: "missing base url",
			raw:  `{"bad":{"task":"generation","provider":"openai_compat","model":"m"}}`,
			want: "base_url is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFromFile(writeModelsConfig(t, tt.raw))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("LoadFromFile() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestLoadFromFileRejectsInvalidBoolFlag(t *testing.T) {
	t.Setenv(string(SkipProviderHealthcheck), "maybe")

	_, err := LoadFromFile(writeModelsConfig(t, `{
		"chat.default": {
			"task": "generation",
			"provider": "openai_compat",
			"model": "chat-model",
			"base_url": "http://provider/v1"
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), string(SkipProviderHealthcheck)) {
		t.Fatalf("LoadFromFile() error = %v, want invalid flag error", err)
	}
}

func writeModelsConfig(t *testing.T, raw string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "models.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write models config: %v", err)
	}

	return path
}
