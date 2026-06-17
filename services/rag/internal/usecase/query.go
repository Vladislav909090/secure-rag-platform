package usecase

import (
	"context"
	"fmt"
	"strings"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"

	"github.com/pgvector/pgvector-go"
)

const (
	// Лимит кандидатов до фильтрации
	contextMaxCandidates = 50
	// Лимит чанков в контексте
	contextMaxChunks = 12
	// Максимальная дистанция от вектора запроса
	contextMaxDistance = 0.8
	// TopK по умолчанию
	defaultTopK = 3
)

const (
	// systemPromptTemplate — системная инструкция для модели
	// Описывает, как отвечать и как ссылаться на источники
	systemPromptTemplate = "Отвечай на русском языке только по предоставленному контексту. " +
		"Если в контексте есть факты, числа, таблицы или CSV-строки, используй их для ответа. " +
		"Если ответа нет в контексте, напиши: Не знаю. " +
		"Отвечай кратко. Не добавляй source IDs и цитирования в текст ответа; " +
		"источники возвращаются отдельно в structured contexts field."
	// userPromptTemplate — шаблон пользовательской части промпта
	userPromptTemplate = "Вопрос:\n%s\n\nКонтекст:\n%s"
)

// Query выполняет поиск контекста и генерацию ответа
func (s *Service) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	question := strings.TrimSpace(req.Query)
	if question == "" {
		return nil, ErrInvalidRequest
	}

	topK := req.TopK
	if topK <= 0 {
		topK = s.defaults.TopK
	}
	if topK <= 0 {
		topK = defaultTopK
	}

	embedAlias := strings.TrimSpace(req.EmbeddingModelAlias)
	if embedAlias == "" {
		embedAlias = s.defaults.EmbeddingModelAlias
	}
	if embedAlias == "" {
		return nil, ErrInvalidRequest
	}

	genAlias := strings.TrimSpace(req.GenerationModelAlias)
	if genAlias == "" {
		genAlias = s.defaults.GenerationModelAlias
	}
	if genAlias == "" {
		return nil, ErrInvalidRequest
	}

	embedResp, err := s.embedding.Embed(ctx, &aiinferencev1.EmbedRequest{
		ModelAlias: embedAlias,
		Text:       question,
		Normalize:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	resolvedModel := strings.TrimSpace(embedResp.GetResolvedModel())
	if resolvedModel == "" {
		resolvedModel = embedAlias
	}
	embeddingDimension := embedResp.GetDimension()
	if embeddingDimension <= 0 {
		embeddingDimension = int32(len(embedResp.GetVector()))
	}
	if embeddingDimension <= 0 || len(embedResp.GetVector()) != int(embeddingDimension) {
		return nil, fmt.Errorf("embedding dimension mismatch")
	}

	candidateLimit := min(max(int(topK)*4, int(topK)), contextMaxCandidates)

	matches, err := s.repo.SearchSimilar(
		ctx,
		pgvector.NewVector(embedResp.GetVector()),
		resolvedModel,
		embeddingDimension,
		s.defaults.IndexedEmbeddingDim,
		int32(candidateLimit),
		req.DocumentUUIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("search chunks: %w", err)
	}

	maxChunks := min(int(topK), contextMaxChunks)

	contexts := make([]QueryContext, 0, len(matches))
	if len(matches) > 0 {
		for _, match := range matches {
			if match.Distance > contextMaxDistance {
				continue
			}
			score := float32(1.0) - match.Distance
			contexts = append(contexts, QueryContext{
				DocumentUUID: match.DocumentUUID,
				ChunkIndex:   match.ChunkIndex,
				Text:         match.ChunkText,
				Score:        score,
			})
			if len(contexts) >= maxChunks {
				break
			}
		}
	}

	if len(contexts) == 0 {
		return &QueryResult{
			Answer:                  "I don't know based on the provided documents.",
			Contexts:                contexts,
			ResolvedEmbeddingModel:  embedResp.GetResolvedModel(),
			ResolvedGenerationModel: "",
		}, nil
	}

	systemPrompt, userPrompt := buildPrompts(question, contexts)

	genResp, err := s.generation.Generate(ctx, &aiinferencev1.GenerateRequest{
		ModelAlias: genAlias,
		Messages: []*aiinferencev1.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("generate answer: %w", err)
	}

	return &QueryResult{
		Answer:                  strings.TrimSpace(genResp.GetContent()),
		Contexts:                contexts,
		ResolvedEmbeddingModel:  embedResp.GetResolvedModel(),
		ResolvedGenerationModel: genResp.GetResolvedModel(),
	}, nil
}

func buildPrompts(question string, contexts []QueryContext) (string, string) {
	parts := make([]string, 0, len(contexts))
	for _, ctx := range contexts {
		parts = append(parts, fmt.Sprintf("[%s:%d]\n%s", ctx.DocumentUUID, ctx.ChunkIndex, ctx.Text))
	}

	system := systemPromptTemplate
	user := fmt.Sprintf(userPromptTemplate, question, strings.Join(parts, "\n\n"))

	return system, user
}
