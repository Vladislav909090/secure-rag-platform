package usecase

import (
	"context"
	"io"
	"strings"
	"testing"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseCreateDocumentUploadsAndStoresMetadata(t *testing.T) {
	t.Parallel()

	const expectedChecksum = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	desc := "description"
	storage := NewMockDocumentStorage(t)
	storage.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything, int64(-1), "application/octet-stream").
		RunAndReturn(func(_ context.Context, key string, reader io.Reader, size int64, contentType string) error {
			data, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.True(t, strings.HasPrefix(key, "documents/"))
			assert.True(t, strings.HasSuffix(key, "/file"))
			assert.Equal(t, int64(-1), size)
			assert.Equal(t, "application/octet-stream", contentType)
			assert.Equal(t, "hello world", string(data))

			return nil
		})
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		CreateDocument(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, doc *model.Document) error {
			require.NotNil(t, doc)
			assert.NotEmpty(t, doc.UUID)
			assert.Equal(t, "title", doc.Title)
			require.NotNil(t, doc.Description)
			assert.Equal(t, desc, *doc.Description)
			assert.Equal(t, map[string]any{}, doc.Attributes)
			assert.Equal(t, "report.pdf", doc.FileName)
			assert.Equal(t, "pdf", doc.FileExtension)
			assert.Equal(t, "application/octet-stream", doc.MimeType)
			assert.Equal(t, int64(11), doc.SizeBytes)
			assert.Equal(t, expectedChecksum, doc.ChecksumSHA256)
			assert.Equal(t, model.IndexStatusPending, doc.IndexStatus)
			assert.NotZero(t, doc.CreatedAt)
			assert.NotZero(t, doc.UpdatedAt)

			return nil
		})
	uc := &DocumentUsecase{repo: repo, storage: storage, maxSize: 64}

	out, err := uc.CreateDocument(context.Background(), CreateDocumentInput{
		Title:       "title",
		Description: &desc,
	}, strings.NewReader("hello world"), "report.pdf", "")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Document)
	assert.Equal(t, "title", out.Document.Title)
}

func TestDocumentUsecaseCreateDocumentRejectsEmptyTitle(t *testing.T) {
	t.Parallel()

	uc := &DocumentUsecase{}

	out, err := uc.CreateDocument(context.Background(), CreateDocumentInput{}, strings.NewReader("hello"), "file.txt", "text/plain")
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, out)
}

func TestDocumentUsecaseCreateDocumentReturnsFileTooLarge(t *testing.T) {
	t.Parallel()

	storage := NewMockDocumentStorage(t)
	storage.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything, int64(-1), "text/plain").
		RunAndReturn(func(_ context.Context, _ string, reader io.Reader, _ int64, _ string) error {
			_, err := io.ReadAll(reader)

			return err
		})
	uc := &DocumentUsecase{
		repo:    NewMockDocumentRepo(t),
		storage: storage,
		maxSize: 3,
	}

	out, err := uc.CreateDocument(context.Background(), CreateDocumentInput{Title: "title"}, strings.NewReader("hello"), "file.txt", "text/plain")
	require.ErrorIs(t, err, ErrFileTooLarge)
	assert.Nil(t, out)
}
