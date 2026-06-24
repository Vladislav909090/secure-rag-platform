package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepo(t *testing.T) {
	t.Parallel()

	repo := NewRepo(nil)
	require.NotNil(t, repo)
	assert.Nil(t, repo.pool)
}

func TestEnsureEmbeddingIndexValidatesDimension(t *testing.T) {
	t.Parallel()

	repo := &Repo{}

	err := repo.EnsureEmbeddingIndex(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "indexed embedding dimension is required")

	err = repo.EnsureEmbeddingIndex(context.Background(), maxIndexedEmbeddingDimension+1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds pgvector vector limit")
}
