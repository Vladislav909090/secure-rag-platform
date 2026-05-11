package main

import (
	"context"
	"testing"

	"secure-rag-platform/services/ai-inference/internal/config"
	transportgrpc "secure-rag-platform/services/ai-inference/internal/transport/grpc"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

type testProvider struct{}

func (testProvider) Name() string {
	return "test"
}

func (testProvider) Generate(
	ctx context.Context,
	alias config.ModelAlias,
	req usecase.GenerateRequest,
) (*usecase.GenerateResult, error) {
	return &usecase.GenerateResult{}, nil
}

func (testProvider) Embed(
	ctx context.Context,
	alias config.ModelAlias,
	req usecase.BatchEmbedRequest,
) (*usecase.BatchEmbedResult, error) {
	return &usecase.BatchEmbedResult{}, nil
}

func TestHealthRPC(t *testing.T) {
	svc := usecase.NewService(
		map[string]config.ModelAlias{
			"chat.default": {
				Task:     config.TaskGeneration,
				Provider: "test",
				Model:    "m",
				BaseURL:  "http://example",
			},
			"embed.default": {
				Task:     config.TaskEmbedding,
				Provider: "test",
				Model:    "m",
				BaseURL:  "http://example",
			},
		},
		[]usecase.Provider{testProvider{}},
		nil,
	)

	server := transportgrpc.NewAIInferenceServiceServer(svc)

	resp, err := server.Health(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if resp.GetStatus() != "ok" {
		t.Fatalf("expected status ok, got %q", resp.GetStatus())
	}
}
