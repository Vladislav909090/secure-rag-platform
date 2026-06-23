package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/repository"
	"secure-rag-platform/services/knowledge/internal/storage"
)

var (
	ErrDocumentNotFound   = errors.New("document not found")
	ErrDocumentDeleted    = errors.New("document is deleted")
	ErrDocumentNotDeleted = errors.New("document is not deleted")
	ErrFileNotFound       = errors.New("file not found in storage")
	ErrFileTooLarge       = errors.New("file too large")
	ErrInvalidRequest     = errors.New("invalid request")
)

// DocumentUsecase содержит бизнес-логику работы с документами
type DocumentUsecase struct {
	repo    DocumentRepo
	storage DocumentStorage
	maxSize int64
}

type DocumentRepo interface {
	CreateDocument(context.Context, *model.Document) error
	GetDocumentByUUID(context.Context, string) (*model.Document, error)
	ListActiveDocuments(context.Context) ([]*model.Document, error)
	RestoreDocument(context.Context, string, time.Time) error
	SoftDeleteDocument(context.Context, string, time.Time) error
	UpdateAttributes(context.Context, string, map[string]any, time.Time) error
	UpdateDocument(context.Context, string, *string, *string, time.Time) error
	UpdateIndexStatus(context.Context, int64, string) error
}

type DocumentStorage interface {
	Delete(context.Context, string) error
	Download(context.Context, string) (io.ReadCloser, error)
	Upload(context.Context, string, io.Reader, int64, string) error
}

// NewDocumentUsecase создаёт сценарий работы с документами
func NewDocumentUsecase(repo *repository.Repo, s3 *storage.S3Storage, maxSize int64) *DocumentUsecase {
	return &DocumentUsecase{repo: repo, storage: s3, maxSize: maxSize}
}

type sizeLimitReader struct {
	r      io.Reader
	max    int64
	read   int64
	failed bool
}

func (r *sizeLimitReader) Read(p []byte) (int, error) {
	if r.failed {
		return 0, ErrFileTooLarge
	}

	n, err := r.r.Read(p)
	r.read += int64(n)
	if r.read > r.max {
		r.failed = true
		return n, ErrFileTooLarge
	}

	return n, err
}

func (uc *DocumentUsecase) uploadAndHash(
	ctx context.Context,
	storageKey string,
	file io.Reader,
	mimeType string,
) (int64, string, error) {
	hasher := sha256.New()
	limited := &sizeLimitReader{r: file, max: uc.maxSize}
	stream := io.TeeReader(limited, hasher)

	if err := uc.storage.Upload(ctx, storageKey, stream, -1, mimeType); err != nil {
		if errors.Is(err, ErrFileTooLarge) || limited.failed {
			return 0, "", ErrFileTooLarge
		}

		return 0, "", err
	}

	if limited.failed {
		return 0, "", ErrFileTooLarge
	}

	return limited.read, hex.EncodeToString(hasher.Sum(nil)), nil
}
