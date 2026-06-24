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

func TestDocumentUsecaseUpdateDocumentUpdatesTitleAndDescription(t *testing.T) {
	t.Parallel()

	title := "new title"
	desc := "new description"
	updated := usecaseTestDocument("doc-1")
	updated.Title = title
	updated.Description = &desc
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)
			return usecaseTestDocument("doc-1"), nil
		}).
		Once()
	repo.EXPECT().
		UpdateDocument(mock.Anything, "doc-1", &title, &desc, mock.Anything).
		RunAndReturn(func(_ context.Context, uuid string, gotTitle *string, gotDesc *string, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			require.NotNil(t, gotTitle)
			require.NotNil(t, gotDesc)
			assert.Equal(t, title, *gotTitle)
			assert.Equal(t, desc, *gotDesc)
			assert.NotZero(t, updatedAt)

			return nil
		})
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		Return(updated, nil).
		Once()
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateDocument(context.Background(), "doc-1", &title, &desc)
	require.NoError(t, err)
	assert.Same(t, updated, out)
}

func TestDocumentUsecaseUpdateDocumentRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateDocument(context.Background(), "doc-1", nil, nil)
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, out)
}
