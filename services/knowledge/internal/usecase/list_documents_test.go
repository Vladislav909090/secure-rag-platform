package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentUsecaseListDocumentsWrapsDocuments(t *testing.T) {
	t.Parallel()

	docs := []*model.Document{
		usecaseTestDocument("doc-1"),
		usecaseTestDocument("doc-2"),
	}
	repo := &mockDocumentRepo{
		t: t,
		listActiveDocuments: func(context.Context) ([]*model.Document, error) {
			return docs, nil
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.ListDocuments(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Same(t, docs[0], out[0].Document)
	assert.Same(t, docs[1], out[1].Document)
}

func TestDocumentUsecaseListDocumentsWrapsRepoError(t *testing.T) {
	t.Parallel()

	repo := &mockDocumentRepo{
		t: t,
		listActiveDocuments: func(context.Context) ([]*model.Document, error) {
			return nil, errBoom()
		},
	}
	uc := &DocumentUsecase{repo: repo}

	out, err := uc.ListDocuments(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list documents")
	assert.Nil(t, out)
}
