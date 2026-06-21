package httpupload

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
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
	if got := extension("report.final.pdf"); got != "pdf" {
		t.Fatalf("extension() = %q", got)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", " Bearer token ")
	if got := extractAccessToken(req); got != "token" {
		t.Fatalf("unexpected token %q", got)
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

func newMultipartFieldPart(t *testing.T, name string, value string) *multipart.Part {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partWriter, createErr := writer.CreateFormField(name)
	if createErr != nil {
		t.Fatalf("CreateFormField() error = %v", createErr)
	}
	_, copyErr := io.Copy(partWriter, strings.NewReader(value))
	if copyErr != nil {
		t.Fatalf("write form field: %v", copyErr)
	}
	closeErr := writer.Close()
	if closeErr != nil {
		t.Fatalf("close multipart writer: %v", closeErr)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	part, nextErr := reader.NextPart()
	if nextErr != nil {
		t.Fatalf("NextPart() error = %v", nextErr)
	}

	return part
}
