package httpupload

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeUploadHelpers(t *testing.T) {
	t.Parallel()

	assert.True(t, isMultipart("multipart/form-data; boundary=abc"))
	assert.False(t, isMultipart("application/json"))
	assert.Equal(t, "gz", extension("archive.tar.gz"))
	assert.Equal(t, "2026.pdf", asciiFilenameFallback("Отчет 2026.pdf"))
	assert.Equal(t, "download.bin", asciiFilenameFallback("   "))
}

func TestKnowledgeSendChunksCopiesDataAndStopsOnSendError(t *testing.T) {
	t.Parallel()

	var chunks [][]byte
	err := sendChunks(strings.NewReader("abcdef"), 2, func(chunk []byte) error {
		chunks = append(chunks, chunk)
		chunk[0] = 'X'

		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "XbXdXf", string(bytes.Join(chunks, nil)))

	errSentinel := errors.New("send failed")
	err = sendChunks(strings.NewReader("abcdef"), 2, func(chunk []byte) error {
		return errSentinel
	})
	require.ErrorIs(t, err, errSentinel)
}

func TestKnowledgeReadLimitedPartRejectsOversizedPart(t *testing.T) {
	t.Parallel()

	part := newMultipartFieldPart(t, "title", "abcdef")

	_, err := readLimitedPart(part, 3)
	require.Error(t, err)
}

func TestParseStructAttributes(t *testing.T) {
	t.Parallel()

	attrs, err := parseStructAttributes([]byte(`{"department":"legal","level":2}`))
	require.NoError(t, err)
	level, ok := attrs.AsMap()["level"].(float64)
	require.True(t, ok)
	assert.Equal(t, "legal", attrs.AsMap()["department"])
	assert.Equal(t, float64(2), level)

	attrs, err = parseStructAttributes([]byte(" "))
	require.NoError(t, err)
	assert.Nil(t, attrs)

	_, err = parseStructAttributes([]byte(`{"bad"`))
	require.Error(t, err)
}

func TestWriteDownloadResponse(t *testing.T) {
	t.Parallel()

	handler := New(nil, nil)
	rec := httptest.NewRecorder()
	handler.writeDownloadResponse(rec, &usecase.FileDownload{
		Body:      io.NopCloser(strings.NewReader("file body")),
		FileName:  "Отчет 2026.txt",
		MimeType:  "",
		SizeBytes: 9,
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "9", rec.Header().Get("Content-Length"))
	assert.Equal(t, "file body", rec.Body.String())
	assert.Contains(t, rec.Header().Get("Content-Disposition"), `filename="2026.txt"`)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "filename*=UTF-8''")
}

func TestDocumentFilesFallsThroughForNonFilePaths(t *testing.T) {
	t.Parallel()

	fallbackCalled := false
	handler := New(nil, nil).DocumentFiles(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusAccepted)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/knowledge/api/v1/documents/doc-1", nil)
	handler(rec, req)

	assert.True(t, fallbackCalled)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func newMultipartFieldPart(t *testing.T, name string, value string) *multipart.Part {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partWriter, err := writer.CreateFormField(name)
	require.NoError(t, err)
	_, err = io.Copy(partWriter, strings.NewReader(value))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	reader := multipart.NewReader(&body, writer.Boundary())
	part, nextErr := reader.NextPart()
	require.NoError(t, nextErr)

	return part
}
