package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseReindexDocumentMarksPending(t *testing.T) {
	t.Parallel()

	doc := usecaseTestDocument("doc-1")
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)

			return doc, nil
		})
	repo.EXPECT().
		UpdateIndexStatus(mock.Anything, doc.ID, model.IndexStatusPending).
		RunAndReturn(func(_ context.Context, docID int64, status string) error {
			assert.Equal(t, doc.ID, docID)
			assert.Equal(t, model.IndexStatusPending, status)

			return nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.ReindexDocument(context.Background(), "doc-1")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "doc-1", out.DocumentUUID)
	assert.Equal(t, model.IndexStatusPending, out.IndexStatus)
}

func TestDocumentUsecaseReindexDocumentPropagatesStatusError(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		})
	repo.EXPECT().
		UpdateIndexStatus(mock.Anything, int64(7), model.IndexStatusPending).
		RunAndReturn(func(context.Context, int64, string) error {
			return errBoom()
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.ReindexDocument(context.Background(), "doc-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update index status")
	assert.Nil(t, out)
}
