package usecase

type IndexDocumentVersionRequest struct {
	DocumentUUID        string
	VersionNumber       int32
	EmbeddingModelAlias string
	ChunkSize           int
	ChunkOverlap        int
}

type IndexDocumentVersionResult struct {
	DocumentUUID           string
	VersionNumber          int32
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
	DocumentUUID  string
	VersionNumber int32
	ChunkIndex    int32
	Text          string
	Score         float32
}

type QueryResult struct {
	Answer                  string
	Contexts                []QueryContext
	ResolvedEmbeddingModel  string
	ResolvedGenerationModel string
}
