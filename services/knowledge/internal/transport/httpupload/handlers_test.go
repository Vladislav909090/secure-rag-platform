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
)

func TestUploadHelpers(t *testing.T) {
	if !isMultipart("multipart/form-data; boundary=abc") {
		t.Fatalf("expected multipart content type")
	}
	if isMultipart("application/json") {
		t.Fatalf("did not expect json to be multipart")
	}
	if got := extension("archive.tar.gz"); got != "gz" {
		t.Fatalf("extension() = %q", got)
	}
	if got := asciiFilenameFallback("Отчет 2026.pdf"); got != "2026.pdf" {
		t.Fatalf("asciiFilenameFallback() = %q", got)
	}
	if got := asciiFilenameFallback("   "); got != "download.bin" {
		t.Fatalf("blank fallback = %q", got)
	}
}

func TestSendChunksCopiesDataAndStopsOnSendError(t *testing.T) {
	var chunks [][]byte
	err := sendChunks(strings.NewReader("abcdef"), 2, func(chunk []byte) error {
		chunks = append(chunks, chunk)
		chunk[0] = 'X'

		return nil
	})
	if err != nil {
		t.Fatalf("sendChunks() error = %v", err)
	}
	if got := string(bytes.Join(chunks, nil)); got != "XbXdXf" {
		t.Fatalf("unexpected mutated local chunks %q", got)
	}

	errSentinel := errors.New("send failed")
	err = sendChunks(strings.NewReader("abcdef"), 2, func(chunk []byte) error {
		return errSentinel
	})
	if !errors.Is(err, errSentinel) {
		t.Fatalf("sendChunks() error = %v, want sentinel", err)
	}
}

func TestReadLimitedPartRejectsOversizedPart(t *testing.T) {
	part := newMultipartFieldPart(t, "title", "abcdef")

	_, err := readLimitedPart(part, 3)
	if err == nil {
		t.Fatalf("expected size limit error")
	}
}

func TestDocumentFilesFallsThroughForNonFilePaths(t *testing.T) {
	fallbackCalled := false
	handler := New(nil, nil).DocumentFiles(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusAccepted)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/knowledge/api/v1/documents/doc-1", nil)
	handler(rec, req)

	if !fallbackCalled || rec.Code != http.StatusAccepted {
		t.Fatalf("expected fallback handler, called=%v status=%d", fallbackCalled, rec.Code)
	}
}

func newMultipartFieldPart(t *testing.T, name string, value string) *multipart.Part {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partWriter, err := writer.CreateFormField(name)
	if err != nil {
		t.Fatalf("CreateFormField() error = %v", err)
	}
	_, err = io.Copy(partWriter, strings.NewReader(value))
	if err != nil {
		t.Fatalf("write form field: %v", err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	part, nextErr := reader.NextPart()
	if nextErr != nil {
		t.Fatalf("NextPart() error = %v", nextErr)
	}

	return part
}
