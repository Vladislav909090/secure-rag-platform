package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRAGServiceDeleteDocumentIndexDeletesChunks(t *testing.T) {
	t.Parallel()

	repo := NewMockRAGRepo(t)
	repo.EXPECT().
		DeleteChunks(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, documentUUID string) error {
			assert.Equal(t, "doc-1", documentUUID)

			return nil
		})

	err := newRAGTestService(t, repo).DeleteDocumentIndex(context.Background(), " doc-1 ")
	require.NoError(t, err)
}

func TestRAGServiceDeleteDocumentIndexRejectsEmptyDocumentUUID(t *testing.T) {
	t.Parallel()

	err := newRAGTestService(t, NewMockRAGRepo(t)).DeleteDocumentIndex(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func TestRAGServiceDeleteDocumentIndexRequiresConfiguredService(t *testing.T) {
	t.Parallel()

	err := (&Service{}).DeleteDocumentIndex(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrNotConfigured)
}
