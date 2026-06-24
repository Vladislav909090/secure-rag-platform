package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

func TestOpenAICompatProviderGenerateAndEmbed(t *testing.T) {
	var seenAuth string
	var seenPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		seenPaths = append(seenPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/chat/completions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{
					"message":       map[string]any{"role": "assistant", "content": "hello"},
					"finish_reason": "stop",
				}},
				"usage": map[string]any{"prompt_tokens": 2, "completion_tokens": 3, "total_tokens": 5},
			})
		case "/v1/embeddings":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"index": 1, "embedding": []float32{0, 1}},
					{"index": 0, "embedding": []float32{1, 0}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	p := NewOpenAICompatProvider(time.Second)
	alias := config.ModelAlias{Model: "m", BaseURL: server.URL + "/v1/", APIKey: " token "}

	generated, err := p.Generate(context.Background(), alias, usecase.GenerateRequest{
		Messages: []usecase.Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if generated.Content != "hello" || generated.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected generate result: %#v", generated)
	}

	embedded, err := p.Embed(context.Background(), alias, usecase.BatchEmbedRequest{Texts: []string{"a", "b"}})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if embedded.Dimension != 2 || embedded.Vectors[0][0] != 1 || embedded.Vectors[1][1] != 1 {
		t.Fatalf("unexpected embedding result: %#v", embedded)
	}
	if seenAuth != "Bearer token" {
		t.Fatalf("unexpected auth header %q", seenAuth)
	}
	if strings.Join(seenPaths, ",") != "/v1/chat/completions,/v1/embeddings" {
		t.Fatalf("unexpected paths: %v", seenPaths)
	}
}

func TestOpenAICompatProviderReturnsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "provider down", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewOpenAICompatProvider(time.Second).Generate(
		context.Background(),
		config.ModelAlias{Model: "m", BaseURL: server.URL},
		usecase.GenerateRequest{Messages: []usecase.Message{{Role: "user", Content: "hello"}}},
	)
	if err == nil || !strings.Contains(err.Error(), "provider status 502") {
		t.Fatalf("Generate() error = %v, want provider status", err)
	}
}

func TestMockProvider(t *testing.T) {
	p := NewMockProvider()
	if p.Name() != config.OpenAICompatProviderName {
		t.Fatalf("unexpected provider name %q", p.Name())
	}

	generated, err := p.Generate(context.Background(), config.ModelAlias{Model: "mock-model"}, usecase.GenerateRequest{
		Messages: []usecase.Message{{Role: "user", Content: "question"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(generated.Content, "mock-model") || !strings.Contains(generated.Content, "question") {
		t.Fatalf("unexpected mock content %q", generated.Content)
	}

	embedded, err := p.Embed(context.Background(), config.ModelAlias{}, usecase.BatchEmbedRequest{Texts: []string{"a", "b"}})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if embedded.Dimension != mockEmbeddingDimension || len(embedded.Vectors) != 2 {
		t.Fatalf("unexpected mock embeddings: %#v", embedded)
	}
}
