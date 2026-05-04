package usecase

type IndexDocumentRequest struct {
	DocumentUUID        string
	EmbeddingModelAlias string
	ChunkSize           int
	ChunkOverlap        int
}

type IndexDocumentResult struct {
	DocumentUUID           string
	ChunkCount             int
	EmbeddingDimension     int32
	ResolvedEmbeddingModel string
}

type QueryRequest struct {
	Query                string
	TopK                 int32
	DocumentUUIDs        []string
	EmbeddingModelAlias  string
	GenerationModelAlias string
}

type QueryContext struct {
	DocumentUUID string
	ChunkIndex   int32
	Text         string
	Score        float32
}

type QueryResult struct {
	Answer                  string
	Contexts                []QueryContext
	ResolvedEmbeddingModel  string
	ResolvedGenerationModel string
}
