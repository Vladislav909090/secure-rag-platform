package usecase

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

type ReindexRequest struct {
	DocumentUUID        string
	EmbeddingModelAlias string
	ChunkSize           int32
	ChunkOverlap        int32
}

type ReindexResult struct {
	DocumentUUID           string
	ChunkCount             int32
	EmbeddingDimension     int32
	ResolvedEmbeddingModel string
}

type ReindexItemResult struct {
	DocumentUUID           string
	Indexed                bool
	Error                  string
	ChunkCount             int32
	EmbeddingDimension     int32
	ResolvedEmbeddingModel string
}

type ReindexAllResult struct {
	TotalCount   int32
	IndexedCount int32
	FailedCount  int32
	Items        []ReindexItemResult
}

type LoginRequest struct {
	Login    string
	Password string
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
}

type SubjectContext struct {
	UserID     string
	Login      string
	IsActive   bool
	Roles      []string
	Attributes map[string]any
	CtxVer     int64
}

type Document struct {
	ID             int64
	UUID           string
	Title          string
	Description    string
	Attributes     map[string]any
	FileName       string
	FileExtension  string
	MimeType       string
	SizeBytes      int64
	ChecksumSHA256 string
	StorageKey     string
	IndexStatus    string
	CreatedAt      string
	UpdatedAt      string
	DeletedAt      string
}

type DocumentItem struct {
	Document Document
}

type FileResult struct {
	ContentType string
	Data        []byte
}

type DeleteDocumentResult struct {
	DocumentUUID string
	Deleted      bool
	DeletedAt    string
	IndexDeleted bool
}
