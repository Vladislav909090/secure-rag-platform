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

func TestDocumentUsecaseGetDocumentReturnsActiveDocument(t *testing.T) {
	t.Parallel()

	doc := usecaseTestDocument("doc-1")
	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "doc-1").
		RunAndReturn(func(_ context.Context, uuid string) (*model.Document, error) {
			assert.Equal(t, "doc-1", uuid)

			return doc, nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.GetDocument(context.Background(), "doc-1")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Same(t, doc, out.Document)
}

func TestDocumentUsecaseGetDocumentReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewMockDocumentRepo(t)
	repo.EXPECT().
		GetDocumentByUUID(mock.Anything, "missing").
		RunAndReturn(func(context.Context, string) (*model.Document, error) {
			return nil, nil
		})
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.GetDocument(context.Background(), "missing")
	require.ErrorIs(t, err, ErrDocumentNotFound)
	assert.Nil(t, out)
}

func TestDocumentUsecaseGetDocumentReturnsDeleted(t *testing.T) {
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

	out, err := uc.GetDocument(context.Background(), "doc-1")
	require.ErrorIs(t, err, ErrDocumentDeleted)
	assert.Nil(t, out)
}
