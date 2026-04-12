package usecase

import (
	"io"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateDocumentInput — входные данные для создания документа.
type CreateDocumentInput struct {
	Title       string
	Description *string
	Attributes  map[string]any
}

// CreateDocumentOutput — результат создания документа.
type CreateDocumentOutput struct {
	Document *model.Document
	Version  *model.DocumentVersion
}

// UploadVersionOutput — результат загрузки новой версии.
type UploadVersionOutput struct {
	DocumentID     int64
	DocumentUUID   string
	CurrentVersion int32
	Version        *model.DocumentVersion
}

// DeleteDocumentOutput — результат soft delete.
type DeleteDocumentOutput struct {
	DocumentUUID string
	DeletedAt    time.Time
}

// DocumentWithVersions — документ со всеми версиями.
type DocumentWithVersions struct {
	Document *model.Document
	Versions []*model.DocumentVersion
}

// DocumentVersionDetail — документ с одной версией.
type DocumentVersionDetail struct {
	Document *model.Document
	Version  *model.DocumentVersion
}

// FileDownload — данные для скачивания файла.
type FileDownload struct {
	Body      io.ReadCloser
	FileName  string
	MimeType  string
	SizeBytes int64
}

// ReindexOutput — результат запроса на переиндексацию.
type ReindexOutput struct {
	DocumentUUID  string
	VersionNumber int32
	IndexStatus   string
}
