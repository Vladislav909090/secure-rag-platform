package usecase

import (
	"context"
	"errors"
	"math"
	"testing"

	"secure-rag-platform/services/ai-inference/internal/config"
)

type fakeProvider struct {
	name         string
	generateErr  error
	embedErr     error
	lastGenerate GenerateRequest
	lastEmbed    BatchEmbedRequest
}

func (p *fakeProvider) Name() string {
	if p.name == "" {
		return config.OpenAICompatProviderName
	}

	return p.name
}

func (p *fakeProvider) Generate(ctx context.Context, alias config.ModelAlias, req GenerateRequest) (*GenerateResult, error) {
	_ = ctx
	_ = alias
	p.lastGenerate = req
	if p.generateErr != nil {
		return nil, p.generateErr
	}

	return &GenerateResult{Content: "ok"}, nil
}

func (p *fakeProvider) Embed(ctx context.Context, alias config.ModelAlias, req BatchEmbedRequest) (*BatchEmbedResult, error) {
	_ = ctx
	_ = alias
	p.lastEmbed = req
	if p.embedErr != nil {
		return nil, p.embedErr
	}

	return &BatchEmbedResult{Vectors: [][]float32{{3, 4}}, Dimension: 2}, nil
}

func TestGenerateMergesDefaultsAndSetsResolvedProvider(t *testing.T) {
	provider := &fakeProvider{}
	defaultTemp := float32(0.3)
	defaultMaxTokens := int32(64)
	incomingTemp := float32(0.8)
	svc := NewService(map[string]config.ModelAlias{
		"chat.default": {
			Task:     config.TaskGeneration,
			Provider: config.OpenAICompatProviderName,
			Model:    "chat-model",
			BaseURL:  "http://provider/v1",
			GenerationDefaults: config.GenerationDefaults{
				Temperature: &defaultTemp,
				MaxTokens:   &defaultMaxTokens,
			},
		},
	}, []Provider{provider}, nil)

	result, err := svc.Generate(context.Background(), GenerateRequest{
		ModelAlias: "chat.default",
		Messages:   []Message{{Role: "user", Content: "hello"}},
		Params:     GenerationParams{Temperature: &incomingTemp},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.ResolvedProvider != config.OpenAICompatProviderName || result.ResolvedModel != "chat-model" {
		t.Fatalf("unexpected resolved provider/model: %#v", result)
	}
	if got := *provider.lastGenerate.Params.Temperature; got != incomingTemp {
		t.Fatalf("expected incoming temperature %.1f, got %.1f", incomingTemp, got)
	}
	if got := *provider.lastGenerate.Params.MaxTokens; got != defaultMaxTokens {
		t.Fatalf("expected default max tokens %d, got %d", defaultMaxTokens, got)
	}
}

func TestGenerateValidationAndProviderErrors(t *testing.T) {
	svc := NewService(map[string]config.ModelAlias{
		"chat.default":  {Task: config.TaskGeneration, Provider: config.OpenAICompatProviderName, Model: "m", BaseURL: "http://x"},
		"embed.default": {Task: config.TaskEmbedding, Provider: config.OpenAICompatProviderName, Model: "m", BaseURL: "http://x"},
	}, []Provider{&fakeProvider{generateErr: errors.New("down")}}, nil)

	if _, err := svc.Generate(context.Background(), GenerateRequest{}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
	if _, err := svc.Generate(context.Background(), GenerateRequest{
		ModelAlias: "missing",
		Messages:   []Message{{Role: "user", Content: "hello"}},
	}); !errors.Is(err, ErrAliasNotFound) {
		t.Fatalf("expected alias not found, got %v", err)
	}
	if _, err := svc.Generate(context.Background(), GenerateRequest{
		ModelAlias: "embed.default",
		Messages:   []Message{{Role: "user", Content: "hello"}},
	}); !errors.Is(err, ErrAliasTaskMismatch) {
		t.Fatalf("expected alias task mismatch, got %v", err)
	}
	if _, err := svc.Generate(context.Background(), GenerateRequest{
		ModelAlias: "chat.default",
		Messages:   []Message{{Role: "user", Content: "hello"}},
	}); !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("expected provider failed, got %v", err)
	}
}

func TestEmbedNormalizesVectorsAndCheckDependenciesCanBeDisabled(t *testing.T) {
	provider := &fakeProvider{generateErr: errors.New("should not be called"), embedErr: errors.New("should not be called")}
	svc := NewService(map[string]config.ModelAlias{
		"chat.default":  {Task: config.TaskGeneration, Provider: config.OpenAICompatProviderName, Model: "chat", BaseURL: "http://x"},
		"embed.default": {Task: config.TaskEmbedding, Provider: config.OpenAICompatProviderName, Model: "embed", BaseURL: "http://x"},
	}, []Provider{provider}, nil)
	svc.SetDependencyChecksDisabled(true)

	if err := svc.CheckDependencies(context.Background()); err != nil {
		t.Fatalf("CheckDependencies() error = %v", err)
	}

	provider.embedErr = nil
	result, err := svc.Embed(context.Background(), BatchEmbedRequest{
		ModelAlias: "embed.default",
		Texts:      []string{"hello"},
		Normalize:  true,
	})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	got := result.Vectors[0]
	if math.Abs(float64(got[0]-0.6)) > 0.0001 || math.Abs(float64(got[1]-0.8)) > 0.0001 {
		t.Fatalf("expected normalized vector [0.6 0.8], got %v", got)
	}
	if result.ResolvedProvider != config.OpenAICompatProviderName || result.ResolvedModel != "embed" {
		t.Fatalf("unexpected resolved fields: %#v", result)
	}
}

func TestListModelsFiltersAndSorts(t *testing.T) {
	svc := NewService(map[string]config.ModelAlias{
		"z.embed": {Task: config.TaskEmbedding, Provider: config.OpenAICompatProviderName, Model: "e", BaseURL: "http://x"},
		"a.chat":  {Task: config.TaskGeneration, Provider: config.OpenAICompatProviderName, Model: "g", BaseURL: "http://x"},
	}, []Provider{&fakeProvider{}}, nil)

	all := svc.ListModels("")
	if len(all) != 2 || all[0].Alias != "a.chat" || all[1].Alias != "z.embed" {
		t.Fatalf("unexpected sorted models: %#v", all)
	}

	embeddings := svc.ListModels(config.TaskEmbedding)
	if len(embeddings) != 1 || embeddings[0].Alias != "z.embed" {
		t.Fatalf("unexpected embedding models: %#v", embeddings)
	}
}
