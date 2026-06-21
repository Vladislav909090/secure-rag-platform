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
)

func TestKnowledgeUploadHelpers(t *testing.T) {
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

func TestKnowledgeSendChunksCopiesDataAndStopsOnSendError(t *testing.T) {
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

func TestKnowledgeReadLimitedPartRejectsOversizedPart(t *testing.T) {
	part := newMultipartFieldPart(t, "title", "abcdef")

	_, err := readLimitedPart(part, 3)
	if err == nil {
		t.Fatalf("expected size limit error")
	}
}

func TestParseStructAttributes(t *testing.T) {
	attrs, err := parseStructAttributes([]byte(`{"department":"legal","level":2}`))
	if err != nil {
		t.Fatalf("parseStructAttributes() error = %v", err)
	}
	level, ok := attrs.AsMap()["level"].(float64)
	if !ok || attrs.AsMap()["department"] != "legal" || level != 2 {
		t.Fatalf("unexpected attrs: %#v", attrs.AsMap())
	}

	attrs, err = parseStructAttributes([]byte(" "))
	if err != nil {
		t.Fatalf("blank attributes error = %v", err)
	}
	if attrs != nil {
		t.Fatalf("blank attributes should produce nil struct, got %#v", attrs)
	}

	if _, err = parseStructAttributes([]byte(`{"bad"`)); err == nil {
		t.Fatalf("expected invalid JSON error")
	}
}

func TestWriteDownloadResponse(t *testing.T) {
	handler := New(nil, nil)
	rec := httptest.NewRecorder()
	handler.writeDownloadResponse(rec, &usecase.FileDownload{
		Body:      io.NopCloser(strings.NewReader("file body")),
		FileName:  "Отчет 2026.txt",
		MimeType:  "",
		SizeBytes: 9,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("content type = %q", got)
	}
	if got := rec.Header().Get("Content-Length"); got != "9" {
		t.Fatalf("content length = %q", got)
	}
	if got := rec.Body.String(); got != "file body" {
		t.Fatalf("body = %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, `filename="2026.txt"`) || !strings.Contains(got, "filename*=UTF-8''") {
		t.Fatalf("unexpected content disposition: %q", got)
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
