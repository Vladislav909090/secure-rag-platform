package provider

import (
	"context"
	"fmt"
	"strings"

	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

const mockEmbeddingDimension = 8

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return config.OpenAICompatProviderName
}

func (p *MockProvider) Generate(
	ctx context.Context,
	alias config.ModelAlias,
	req usecase.GenerateRequest,
) (*usecase.GenerateResult, error) {
	_ = ctx
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = strings.TrimSpace(req.Messages[len(req.Messages)-1].Content)
	}
	if prompt == "" {
		prompt = "empty prompt"
	}

	content := fmt.Sprintf("[mock:%s] %s", alias.Model, prompt)

	return &usecase.GenerateResult{
		Content:      content,
		FinishReason: "stop",
		Usage: usecase.Usage{
			PromptTokens:     int32(len(req.Messages)),
			CompletionTokens: int32(len(strings.Fields(content))),
			TotalTokens:      int32(len(req.Messages) + len(strings.Fields(content))),
		},
	}, nil
}

func (p *MockProvider) Embed(
	ctx context.Context,
	alias config.ModelAlias,
	req usecase.BatchEmbedRequest,
) (*usecase.BatchEmbedResult, error) {
	_ = ctx
	_ = alias

	vectors := make([][]float32, 0, len(req.Texts))
	for i, text := range req.Texts {
		vectors = append(vectors, mockVector(text, i))
	}

	return &usecase.BatchEmbedResult{
		Vectors:   vectors,
		Dimension: mockEmbeddingDimension,
	}, nil
}

func mockVector(text string, offset int) []float32 {
	vector := make([]float32, mockEmbeddingDimension)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		trimmed = "mock"
	}

	for i, r := range trimmed {
		idx := (i + offset) % mockEmbeddingDimension
		vector[idx] += float32((int(r)%17)+1) / 17.0
	}

	return vector
}
