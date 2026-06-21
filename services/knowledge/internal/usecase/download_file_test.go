package usecase

import (
	"context"
	"io"
	"strings"
	"testing"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseDownloadFileReturnsBodyAndMetadata(t *testing.T) {
	t.Parallel()

	doc := usecaseTestDocument("doc-1")
	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)

			return doc, nil
		},
	}
	storage := &mockDocumentStorage{
		t: t,
		download: func(_ context.Context, key string) (io.ReadCloser, error) {
			assert.Equal(t, doc.StorageKey, key)

			return io.NopCloser(strings.NewReader("file body")), nil
		},
	}
	uc := &DocumentUsecase{repo: repo, storage: storage}

	out, err := uc.DownloadFile(context.Background(), "doc-1")
	require.NoError(t, err)
	require.NotNil(t, out)
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	require.NoError(t, err)
	assert.Equal(t, "file body", string(body))
	assert.Equal(t, doc.FileName, out.FileName)
	assert.Equal(t, doc.MimeType, out.MimeType)
	assert.Equal(t, doc.SizeBytes, out.SizeBytes)
}

func TestDocumentUsecaseDownloadFileMapsStorageError(t *testing.T) {
	t.Parallel()

	doc := usecaseTestDocument("doc-1")
	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(context.Context, string) (*model.Document, error) {
			return doc, nil
		},
	}
	storage := &mockDocumentStorage{
		t: t,
		download: func(context.Context, string) (io.ReadCloser, error) {
			return nil, errBoom()
		},
	}
	uc := &DocumentUsecase{repo: repo, storage: storage}

	out, err := uc.DownloadFile(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrFileNotFound)
	assert.Nil(t, out)
}
