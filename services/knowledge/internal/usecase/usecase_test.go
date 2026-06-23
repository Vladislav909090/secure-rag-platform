package usecase

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
