package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseDeleteDocumentSoftDeletesActiveDocument(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)

			return usecaseTestDocument("doc-1"), nil
		})
	repo.EXPECT().
		SoftDeleteDocument(mock.Anything, "doc-1", mock.Anything).
		RunAndReturn(func(_ context.Context, uuid string, deletedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			assert.NotZero(t, deletedAt)

			return nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.DeleteDocument(context.Background(), "doc-1")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "doc-1", out.DocumentUUID)
	assert.NotZero(t, out.DeletedAt)
}

func TestDocumentUsecaseDeleteDocumentRejectsDeletedDocument(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now().UTC()
	doc := usecaseTestDocument("doc-1")
	doc.DeletedAt = &deletedAt
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return doc, nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.DeleteDocument(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrDocumentDeleted)
	assert.Nil(t, out)
}
