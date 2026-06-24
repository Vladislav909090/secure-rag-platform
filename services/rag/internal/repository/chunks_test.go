package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSimilarReturnsBeforeDatabaseForEmptyInputs(t *testing.T) {
	t.Parallel()

	repo := &Repo{}

	got, err := repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 1, 0, 0, []string{"doc"})
	require.NoError(t, err)
	assert.Nil(t, got)

	got, err = repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 1, 0, 3, nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSearchSimilarValidatesModelAndDimension(t *testing.T) {
	t.Parallel()

	repo := &Repo{}

	_, err := repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), " ", 1, 0, 3, []string{"doc"})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "embedding model is required"))

	_, err = repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 0, 0, 3, []string{"doc"})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "embedding dimension is required"))
}
