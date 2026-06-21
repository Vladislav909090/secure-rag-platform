package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/pgvector/pgvector-go"
)

func TestSearchSimilarReturnsBeforeDatabaseForEmptyInputs(t *testing.T) {
	repo := &Repo{}

	got, err := repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 1, 0, 0, []string{"doc"})
	if err != nil {
		t.Fatalf("SearchSimilar() error = %v", err)
	}
	if got != nil {
		t.Fatalf("topK <= 0 should return nil matches, got %v", got)
	}

	got, err = repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 1, 0, 3, nil)
	if err != nil {
		t.Fatalf("SearchSimilar() error = %v", err)
	}
	if got != nil {
		t.Fatalf("empty document filter should return nil matches, got %v", got)
	}
}

func TestSearchSimilarValidatesModelAndDimension(t *testing.T) {
	repo := &Repo{}

	_, err := repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), " ", 1, 0, 3, []string{"doc"})
	if err == nil || !strings.Contains(err.Error(), "embedding model is required") {
		t.Fatalf("SearchSimilar() error = %v, want model validation", err)
	}

	_, err = repo.SearchSimilar(context.Background(), pgvector.NewVector([]float32{1}), "model", 0, 0, 3, []string{"doc"})
	if err == nil || !strings.Contains(err.Error(), "embedding dimension is required") {
		t.Fatalf("SearchSimilar() error = %v, want dimension validation", err)
	}
}
