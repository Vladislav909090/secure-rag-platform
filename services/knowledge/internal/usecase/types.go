package usecase

import (
	"io"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
)

// CreateDocumentInput — входные данные для создания документа
type CreateDocumentInput struct {
	Title       string
	Description *string
	Attributes  map[string]any
}

// CreateDocumentOutput — результат создания документа
type CreateDocumentOutput struct {
	Document *model.Document
}

// DeleteDocumentOutput — результат мягкого удаления
type DeleteDocumentOutput struct {
	DocumentUUID string
	DeletedAt    time.Time
}

// DocumentDetail — документ с метаданными файла
type DocumentDetail struct {
	Document *model.Document
}

// FileDownload — данные для скачивания файла
type FileDownload struct {
	Body      io.ReadCloser
	FileName  string
	MimeType  string
	SizeBytes int64
}

// ReindexOutput — результат запроса на переиндексацию
type ReindexOutput struct {
	DocumentUUID string
	IndexStatus  string
}
