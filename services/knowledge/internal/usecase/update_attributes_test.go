package usecase

import (
	"context"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseUpdateAttributesReplacesAttributes(t *testing.T) {
	t.Parallel()

	updated := usecaseTestDocument("doc-1")
	updated.Attributes = map[string]any{"team": "rag"}
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
		updateAttributes: func(_ context.Context, uuid string, attrs map[string]any, updatedAt time.Time) error {
			assert.Equal(t, "doc-1", uuid)
			assert.Equal(t, map[string]any{"team": "rag"}, attrs)
			assert.NotZero(t, updatedAt)

			return nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateAttributes(context.Background(), "doc-1", map[string]any{"team": "rag"})
	require.NoError(t, err)
	assert.Same(t, updated, out)
	assert.Equal(t, 2, getCalls)
}

func TestDocumentUsecaseUpdateAttributesRejectsNilAttributes(t *testing.T) {
	t.Parallel()

	repo := &mockDocumentRepo{
		t: t,
		getDocumentByUUID: func(context.Context, string) (*model.Document, error) {
			return usecaseTestDocument("doc-1"), nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.UpdateAttributes(context.Background(), "doc-1", nil)
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, out)
}
