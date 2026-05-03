package usecase

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

type IndexRequest struct {
	DocumentUUID        string
	VersionNumber       int32
	EmbeddingModelAlias string
	ChunkSize           int32
	ChunkOverlap        int32
}

type IndexResult struct {
	DocumentUUID           string
	VersionNumber          int32
	ChunkCount             int32
	EmbeddingDimension     int32
	ResolvedEmbeddingModel string
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
	ID                   int64
	UUID                 string
	Title                string
	Description          string
	Attributes           map[string]any
	CurrentVersionNumber int32
	CreatedAt            string
	UpdatedAt            string
	DeletedAt            string
}

type DocumentVersion struct {
	ID             int64
	UUID           string
	DocumentID     int64
	VersionNumber  int32
	FileName       string
	FileExtension  string
	MimeType       string
	SizeBytes      int64
	ChecksumSHA256 string
	StorageKey     string
	IndexStatus    string
	CreatedAt      string
}

type DocumentWithVersions struct {
	Document Document
	Versions []DocumentVersion
}

type FileResult struct {
	ContentType string
	Data        []byte
}
