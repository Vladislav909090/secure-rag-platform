package usecase

import (
	"context"

	"secure-rag-platform/services/ai-inference/internal/config"
)

type Message struct {
	Role    string
	Content string
}

type GenerationParams struct {
	Temperature      *float32
	TopP             *float32
	MaxTokens        *int32
	PresencePenalty  *float32
	FrequencyPenalty *float32
}

type GenerateRequest struct {
	RequestID  string
	ModelAlias string
	Messages   []Message
	Params     GenerationParams
}

type Usage struct {
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

type GenerateResult struct {
	Content          string
	FinishReason     string
	Usage            Usage
	ResolvedProvider string
	ResolvedModel    string
}

type EmbeddingInputType string

const (
	EmbeddingInputTypeUnspecified EmbeddingInputType = "UNSPECIFIED"
	EmbeddingInputTypeQuery       EmbeddingInputType = "QUERY"
	EmbeddingInputTypeDocument    EmbeddingInputType = "DOCUMENT"
)

type BatchEmbedRequest struct {
	RequestID  string
	ModelAlias string
	Texts      []string
	InputType  EmbeddingInputType
	Normalize  bool
}

type EmbedResult struct {
	Vector           []float32
	Dimension        int32
	ResolvedProvider string
	ResolvedModel    string
}

type BatchEmbedResult struct {
	Vectors          [][]float32
	Dimension        int32
	ResolvedProvider string
	ResolvedModel    string
}

type ModelInfo struct {
	Alias    string
	Task     config.TaskType
	Provider string
	Model    string
	BaseURL  string
}

type Provider interface {
	Name() string
	Generate(ctx context.Context, alias config.ModelAlias, req GenerateRequest) (*GenerateResult, error)
	Embed(ctx context.Context, alias config.ModelAlias, req BatchEmbedRequest) (*BatchEmbedResult, error)
}
