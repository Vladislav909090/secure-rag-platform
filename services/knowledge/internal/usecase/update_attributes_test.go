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

func TestDocumentUsecaseUpdateAttributesReplacesAttributes(t *testing.T) {
	t.Parallel()

	updated := usecaseTestDocument("doc-1")
	updated.Attributes = map[string]any{"team": "rag"}
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)
			return usecaseTestDocument("doc-1"), nil
		}).
		Once()
	repo.EXPECT().
		UpdateAttributes(mock.Anything, "doc-1", map[string]any{"team": "rag"}, mock.Anything).
		RunAndReturn(func(_ context.Context, uuid string, attrs map[string]any, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			assert.Equal(t, map[string]any{"team": "rag"}, attrs)
			assert.NotZero(t, updatedAt)

			return nil
		})
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		Return(updated, nil).
		Once()
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateAttributes(context.Background(), "doc-1", map[string]any{"team": "rag"})
	require.NoError(t, err)
	assert.Same(t, updated, out)
}

func TestDocumentUsecaseUpdateAttributesRejectsNilAttributes(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateAttributes(context.Background(), "doc-1", nil)
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, out)
}
