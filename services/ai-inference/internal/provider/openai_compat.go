package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

type OpenAICompatProvider struct {
	httpClient *http.Client
}

func NewOpenAICompatProvider(timeout time.Duration) *OpenAICompatProvider {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &OpenAICompatProvider{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (p *OpenAICompatProvider) Name() string {
	return "openai_compat"
}

func (p *OpenAICompatProvider) Generate(ctx context.Context, alias config.ModelAlias, req usecase.GenerateRequest) (*usecase.GenerateResult, error) {
	payload := chatCompletionsRequest{
		Model:            alias.Model,
		Messages:         make([]chatMessage, 0, len(req.Messages)),
		Temperature:      req.Params.Temperature,
		TopP:             req.Params.TopP,
		MaxTokens:        req.Params.MaxTokens,
		PresencePenalty:  req.Params.PresencePenalty,
		FrequencyPenalty: req.Params.FrequencyPenalty,
	}

	for _, msg := range req.Messages {
		payload.Messages = append(payload.Messages, chatMessage{Role: msg.Role, Content: msg.Content})
	}

	body, err := p.postJSON(ctx, alias, "chat/completions", payload)
	if err != nil {
		return nil, err
	}

	var response chatCompletionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode chat response: %w", err)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty chat choices")
	}

	choice := response.Choices[0]
	return &usecase.GenerateResult{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: usecase.Usage{
			PromptTokens:     int32(response.Usage.PromptTokens),
			CompletionTokens: int32(response.Usage.CompletionTokens),
			TotalTokens:      int32(response.Usage.TotalTokens),
		},
	}, nil
}

func (p *OpenAICompatProvider) Embed(ctx context.Context, alias config.ModelAlias, req usecase.BatchEmbedRequest) (*usecase.BatchEmbedResult, error) {
	payload := embeddingsRequest{
		Model: alias.Model,
		Input: req.Texts,
	}

	body, err := p.postJSON(ctx, alias, "embeddings", payload)
	if err != nil {
		return nil, err
	}

	var response embeddingsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode embeddings response: %w", err)
	}
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("empty embeddings data")
	}

	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].Index < response.Data[j].Index
	})

	vectors := make([][]float32, 0, len(response.Data))
	for _, item := range response.Data {
		vector := make([]float32, len(item.Embedding))
		copy(vector, item.Embedding)
		vectors = append(vectors, vector)
	}

	dimension := int32(0)
	if len(vectors) > 0 {
		dimension = int32(len(vectors[0]))
	}

	return &usecase.BatchEmbedResult{
		Vectors:   vectors,
		Dimension: dimension,
	}, nil
}

func (p *OpenAICompatProvider) postJSON(ctx context.Context, alias config.ModelAlias, path string, payload any) ([]byte, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	endpoint := joinURL(alias.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(alias.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(alias.APIKey))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call provider: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read provider response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		const maxBody = 2048
		if len(body) > maxBody {
			body = body[:maxBody]
		}
		return nil, fmt.Errorf("provider status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func joinURL(baseURL string, path string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	tail := strings.TrimLeft(path, "/")
	return base + "/" + tail
}

type chatCompletionsRequest struct {
	Model            string        `json:"model"`
	Messages         []chatMessage `json:"messages"`
	Temperature      *float32      `json:"temperature,omitempty"`
	TopP             *float32      `json:"top_p,omitempty"`
	MaxTokens        *int32        `json:"max_tokens,omitempty"`
	PresencePenalty  *float32      `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32      `json:"frequency_penalty,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []chatChoice `json:"choices"`
	Usage   usagePayload `json:"usage"`
}

type chatChoice struct {
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type embeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingsResponse struct {
	Data []embeddingData `json:"data"`
}

type embeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type usagePayload struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
