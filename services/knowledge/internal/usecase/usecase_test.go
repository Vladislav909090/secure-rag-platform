package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDocumentRepo struct {
	t *testing.T

	createDocument      func(context.Context, *model.Document) error
	getDocumentByUUID   func(context.Context, string) (*model.Document, error)
	listActiveDocuments func(context.Context) ([]*model.Document, error)
	restoreDocument     func(context.Context, string, time.Time) error
	softDeleteDocument  func(context.Context, string, time.Time) error
	updateAttributes    func(context.Context, string, map[string]any, time.Time) error
	updateDocument      func(context.Context, string, *string, *string, time.Time) error
	updateIndexStatus   func(context.Context, int64, string) error
}

var _ documentRepo = (*mockDocumentRepo)(nil)

func (m *mockDocumentRepo) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected repo call", "unexpected repo call: %s", name)
	}
	panic(fmt.Sprintf("unexpected repo call: %s", name))
}

func (m *mockDocumentRepo) CreateDocument(ctx context.Context, doc *model.Document) error {
	if m.createDocument == nil {
		m.unexpected("CreateDocument")
	}

	return m.createDocument(ctx, doc)
}

func (m *mockDocumentRepo) GetDocumentByUUID(ctx context.Context, uuid string) (*model.Document, error) {
	if m.getDocumentByUUID == nil {
		m.unexpected("GetDocumentByUUID")
	}

	return m.getDocumentByUUID(ctx, uuid)
}

func (m *mockDocumentRepo) ListActiveDocuments(ctx context.Context) ([]*model.Document, error) {
	if m.listActiveDocuments == nil {
		m.unexpected("ListActiveDocuments")
	}

	return m.listActiveDocuments(ctx)
}

func (m *mockDocumentRepo) RestoreDocument(ctx context.Context, uuid string, updatedAt time.Time) error {
	if m.restoreDocument == nil {
		m.unexpected("RestoreDocument")
	}

	return m.restoreDocument(ctx, uuid, updatedAt)
}

func (m *mockDocumentRepo) SoftDeleteDocument(ctx context.Context, uuid string, deletedAt time.Time) error {
	if m.softDeleteDocument == nil {
		m.unexpected("SoftDeleteDocument")
	}

	return m.softDeleteDocument(ctx, uuid, deletedAt)
}

func (m *mockDocumentRepo) UpdateAttributes(ctx context.Context, uuid string, attributes map[string]any, updatedAt time.Time) error {
	if m.updateAttributes == nil {
		m.unexpected("UpdateAttributes")
	}

	return m.updateAttributes(ctx, uuid, attributes, updatedAt)
}

func (m *mockDocumentRepo) UpdateDocument(ctx context.Context, uuid string, title *string, description *string, updatedAt time.Time) error {
	if m.updateDocument == nil {
		m.unexpected("UpdateDocument")
	}

	return m.updateDocument(ctx, uuid, title, description, updatedAt)
}

func (m *mockDocumentRepo) UpdateIndexStatus(ctx context.Context, docID int64, status string) error {
	if m.updateIndexStatus == nil {
		m.unexpected("UpdateIndexStatus")
	}

	return m.updateIndexStatus(ctx, docID, status)
}

type mockDocumentStorage struct {
	t *testing.T

	delete   func(context.Context, string) error
	download func(context.Context, string) (io.ReadCloser, error)
	upload   func(context.Context, string, io.Reader, int64, string) error
}

var _ documentStorage = (*mockDocumentStorage)(nil)

func (m *mockDocumentStorage) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected storage call", "unexpected storage call: %s", name)
	}
	panic(fmt.Sprintf("unexpected storage call: %s", name))
}

func (m *mockDocumentStorage) Delete(ctx context.Context, key string) error {
	if m.delete == nil {
		m.unexpected("Delete")
	}

	return m.delete(ctx, key)
}

func (m *mockDocumentStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.download == nil {
		m.unexpected("Download")
	}

	return m.download(ctx, key)
}

func (m *mockDocumentStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if m.upload == nil {
		m.unexpected("Upload")
	}

	return m.upload(ctx, key, reader, size, contentType)
}

func usecaseTestDocument(uuid string) *model.Document {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	desc := "description"

	return &model.Document{
		ID:             7,
		UUID:           uuid,
		Title:          "title",
		Description:    &desc,
		Attributes:     map[string]any{"department": "search"},
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      11,
		ChecksumSHA256: "checksum",
		StorageKey:     "documents/" + uuid + "/file",
		IndexStatus:    model.IndexStatusReady,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func errBoom() error {
	return errors.New("boom")
}

func TestSizeLimitReaderAllowsExactLimit(t *testing.T) {
	t.Parallel()

	reader := &sizeLimitReader{r: strings.NewReader("abc"), max: 3}

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "abc", string(data))
	assert.Equal(t, int64(3), reader.read)
	assert.False(t, reader.failed)
}

func TestSizeLimitReaderRejectsOverflowAndStaysFailed(t *testing.T) {
	t.Parallel()

	reader := &sizeLimitReader{r: strings.NewReader("abcd"), max: 3}
	buf := make([]byte, 4)

	n, err := reader.Read(buf)
	require.ErrorIs(t, err, ErrFileTooLarge)
	assert.Equal(t, 4, n)
	assert.True(t, reader.failed)

	n, err = reader.Read(buf)
	require.ErrorIs(t, err, ErrFileTooLarge)
	assert.Equal(t, 0, n)
}
