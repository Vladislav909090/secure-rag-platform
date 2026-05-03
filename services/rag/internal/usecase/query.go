package usecase

import (
	"context"
	"fmt"
	"strings"

	aiinferencev1 "secure-rag-platform/services/ai-inference/gen/v1"

	"github.com/pgvector/pgvector-go"
)

const (
	// contextMaxCandidates ограничивает число ближайших кандидатов, которые мы забираем из БД до фильтрации.
	contextMaxCandidates = 50
	// contextMaxChunks задаёт жёсткий предел на размер контекста, чтобы не раздувать промпт.
	contextMaxChunks = 12
	// contextMaxDistance отсекает чанки, которые слишком далеко от вектора запроса.
	contextMaxDistance = 0.8
	// contextDistanceEpsilon даёт небольшой запас относительно лучшего расстояния.
	contextDistanceEpsilon = 0.05
	// defaultTopK — значение по умолчанию для TopK, если не задано в запросе или конфиге.
	defaultTopK = 3
)

const (
	// systemPromptTemplate — системная инструкция для модели.
	// Описывает, как отвечать и как ссылаться на источники.
	systemPromptTemplate = "You answer questions using the provided context. " +
		"If the answer is not in the context, say you don't know. " +
		"Keep the answer concise. Do not include source IDs or citations in the answer; " +
		"sources are returned separately in the structured contexts field."
	// userPromptTemplate — шаблон пользовательской части промпта.
	userPromptTemplate = "Question:\n%s\n\nContext:\n%s"
)

// Query выполняет поиск контекста и генерацию ответа.
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

	candidateLimit := max(int(topK)*4, int(topK))
	candidateLimit = min(candidateLimit, contextMaxCandidates)
	candidateLimit = min(candidateLimit, int(topK))

	matches, err := s.repo.SearchSimilar(
		ctx,
		pgvector.NewVector(embedResp.GetVector()),
		resolvedModel,
		int32(candidateLimit),
		req.DocumentUUIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("search chunks: %w", err)
	}

	maxChunks := max(int(topK), contextMaxChunks)

	contexts := make([]QueryContext, 0, len(matches))
	if len(matches) > 0 {
		bestDistance := matches[0].Distance
		allowedDistance := bestDistance + contextDistanceEpsilon
		if allowedDistance > contextMaxDistance {
			allowedDistance = contextMaxDistance
		}

		for _, match := range matches {
			if match.Distance > allowedDistance {
				continue
			}
			score := float32(1.0) - match.Distance
			contexts = append(contexts, QueryContext{
				DocumentUUID:  match.DocumentUUID,
				VersionNumber: match.VersionNumber,
				ChunkIndex:    match.ChunkIndex,
				Text:          match.ChunkText,
				Score:         score,
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
		parts = append(parts, fmt.Sprintf("[%s#%d:%d]\n%s", ctx.DocumentUUID, ctx.VersionNumber, ctx.ChunkIndex, ctx.Text))
	}

	system := systemPromptTemplate
	user := fmt.Sprintf(userPromptTemplate, question, strings.Join(parts, "\n\n"))
	return system, user
}
