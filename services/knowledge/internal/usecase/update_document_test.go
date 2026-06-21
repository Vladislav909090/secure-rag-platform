package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseUpdateDocumentUpdatesTitleAndDescription(t *testing.T) {
	t.Parallel()

	title := "new title"
	desc := "new description"
	updated := usecaseTestDocument("doc-1")
	updated.Title = title
	updated.Description = &desc
	getCalls := 0
	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)
			getCalls++
			if getCalls == 1 {
				return usecaseTestDocument("doc-1"), nil
			}

			return updated, nil
		},
		updateDocument: func(_ context.Context, uuid string, gotTitle *string, gotDesc *string, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			require.NotNil(t, gotTitle)
			require.NotNil(t, gotDesc)
			assert.Equal(t, title, *gotTitle)
			assert.Equal(t, desc, *gotDesc)
			assert.NotZero(t, updatedAt)

			return nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateDocument(context.Background(), "doc-1", &title, &desc)
	require.NoError(t, err)
	assert.Same(t, updated, out)
	assert.Equal(t, 2, getCalls)
}

func TestDocumentUsecaseUpdateDocumentRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateDocument(context.Background(), "doc-1", nil, nil)
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, out)
}
