package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseRestoreDocumentRestoresDeletedDocument(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now().UTC()
	deletedDoc := usecaseTestDocument("doc-1")
	deletedDoc.DeletedAt = &deletedAt
	restoredDoc := usecaseTestDocument("doc-1")
	getCalls := 0
	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)
			getCalls++
			if getCalls == 1 {
				return deletedDoc, nil
			}

			return restoredDoc, nil
		},
		restoreDocument: func(_ context.Context, uuid string, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			assert.NotZero(t, updatedAt)

			return nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.RestoreDocument(context.Background(), "doc-1")
	require.NoError(t, err)
	assert.Same(t, restoredDoc, out)
	assert.Equal(t, 2, getCalls)
}

func TestDocumentUsecaseRestoreDocumentRejectsActiveDocument(t *testing.T) {
	t.Parallel()

	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.RestoreDocument(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrDocumentNotDeleted)
	assert.Nil(t, out)
}
