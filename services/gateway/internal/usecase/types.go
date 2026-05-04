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

type CreateUserRequest struct {
	Login      string
	Password   string
	IsActive   *bool
	RoleCodes  []string
	Attributes map[string]any
}

type UpdateUserRequest struct {
	UserID   string
	Login    *string
	Password *string
	IsActive *bool
}

type User struct {
	ID         string
	Login      string
	IsActive   bool
	CtxVer     int64
	Roles      []string
	Attributes map[string]any
	CreatedAt  string
	UpdatedAt  string
}

type Role struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   string
}

type UserRolesResult struct {
	UserID string
	Roles  []Role
	CtxVer int64
}

type UserAttributesResult struct {
	UserID     string
	Attributes map[string]any
	CtxVer     int64
}

type UpdateDocumentRequest struct {
	DocumentUUID string
	Title        *string
	Description  *string
}

type UpdateDocumentAttributesRequest struct {
	DocumentUUID string
	Attributes   map[string]any
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
