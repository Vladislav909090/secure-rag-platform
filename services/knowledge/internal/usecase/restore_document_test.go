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

func TestDocumentUsecaseRestoreDocumentRestoresDeletedDocument(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now().UTC()
	deletedDoc := usecaseTestDocument("doc-1")
	deletedDoc.DeletedAt = &deletedAt
	restoredDoc := usecaseTestDocument("doc-1")
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)
			return deletedDoc, nil
		}).
		Once()
	repo.EXPECT().
		RestoreDocument(mock.Anything, "doc-1", mock.Anything).
		RunAndReturn(func(_ context.Context, uuid string, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			assert.NotZero(t, updatedAt)

			return nil
		})
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		Return(restoredDoc, nil).
		Once()
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.RestoreDocument(context.Background(), "doc-1")
	require.NoError(t, err)
	assert.Same(t, restoredDoc, out)
}

func TestDocumentUsecaseRestoreDocumentRejectsActiveDocument(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.RestoreDocument(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrDocumentNotDeleted)
	assert.Nil(t, out)
}
