package usecase

import (
	"context"
	"fmt"
	"io"
	"mime"
	"strings"

	aiinferencev1 "secure-rag-platform/services/ai-inference/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/rag/internal/repository"

	"github.com/pgvector/pgvector-go"
)

// IndexDocument индексирует документ в векторное хранилище.
func (s *Service) IndexDocument(
	ctx context.Context,
	req IndexDocumentRequest,
) (*IndexDocumentResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	docUUID := strings.TrimSpace(req.DocumentUUID)
	if docUUID == "" {
		return nil, ErrInvalidRequest
	}

	resp, err := s.knowledge.GetDocument(ctx, &knowledgev1.GetDocumentRequest{DocumentUuid: docUUID})
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	doc := resp.GetDocument()
	if doc == nil {
		return nil, ErrInvalidRequest
	}

	storageKey := strings.TrimSpace(doc.GetStorageKey())
	if storageKey == "" {
		return nil, ErrInvalidRequest
	}

	if !isTextMime(doc.GetMimeType()) {
		return nil, fmt.Errorf("%w: unsupported mime type: %s", ErrInvalidRequest, doc.GetMimeType())
	}

	reader, err := s.storage.Download(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("download from storage: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read storage object: %w", err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, ErrInvalidRequest
	}

	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = s.defaults.ChunkSize
	}
	chunkOverlap := req.ChunkOverlap
	if chunkOverlap <= 0 {
		chunkOverlap = s.defaults.ChunkOverlap
	}

	chunks := chunkText(text, chunkSize, chunkOverlap)
	if len(chunks) == 0 {
		return nil, ErrInvalidRequest
	}

	alias := strings.TrimSpace(req.EmbeddingModelAlias)
	if alias == "" {
		alias = s.defaults.EmbeddingModelAlias
	}
	if alias == "" {
		return nil, ErrInvalidRequest
	}

	embedResp, err := s.embedding.BatchEmbed(ctx, &aiinferencev1.BatchEmbedRequest{
		ModelAlias: alias,
		Texts:      chunks,
		Normalize:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}
	if len(embedResp.GetVectors()) != len(chunks) {
		return nil, fmt.Errorf("embedding size mismatch")
	}

	err = s.repo.DeleteChunks(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("delete previous chunks: %w", err)
	}

	resolvedModel := strings.TrimSpace(embedResp.GetResolvedModel())
	if resolvedModel == "" {
		resolvedModel = alias
	}

	entries := make([]repository.Chunk, 0, len(chunks))
	for i, chunk := range chunks {
		vec := embedResp.GetVectors()[i].GetValues()
		entries = append(entries, repository.Chunk{
			DocumentUUID:       docUUID,
			ChunkIndex:         int32(i),
			ChunkText:          chunk,
			Embedding:          pgvector.NewVector(vec),
			EmbeddingModel:     resolvedModel,
			EmbeddingDimension: embedResp.GetDimension(),
		})
	}

	err = s.repo.InsertChunks(ctx, entries)
	if err != nil {
		return nil, fmt.Errorf("insert chunks: %w", err)
	}

	s.logger.InfoContext(ctx, "документ проиндексирован",
		"component", "rag.index",
		"document_uuid", docUUID,
		"chunks", len(chunks),
	)

	return &IndexDocumentResult{
		DocumentUUID:           docUUID,
		ChunkCount:             len(chunks),
		EmbeddingDimension:     embedResp.GetDimension(),
		ResolvedEmbeddingModel: resolvedModel,
	}, nil
}

func isTextMime(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		mediaType = value
	}

	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	switch mediaType {
	case "application/json", "application/xml", "application/markdown", "application/xhtml+xml":
		return true
	default:
		return false
	}
}
