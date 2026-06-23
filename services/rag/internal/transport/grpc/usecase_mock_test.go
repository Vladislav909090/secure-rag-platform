package grpc

import (
	"context"
	"fmt"
	"testing"

	"secure-rag-platform/services/rag/internal/usecase"

	"github.com/stretchr/testify/require"
)

type mockRAGUsecase struct {
	t *testing.T

	deleteDocumentIndex func(context.Context, string) error
	indexDocument       func(context.Context, usecase.IndexDocumentRequest) (*usecase.IndexDocumentResult, error)
	query               func(context.Context, usecase.QueryRequest) (*usecase.QueryResult, error)
	ready               func() bool
}

var _ ragUsecase = (*mockRAGUsecase)(nil)

func (m *mockRAGUsecase) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected usecase call", "unexpected usecase call: %s", name)
	}
	panic(fmt.Sprintf("unexpected usecase call: %s", name))
}

func (m *mockRAGUsecase) DeleteDocumentIndex(ctx context.Context, documentUUID string) error {
	if m.deleteDocumentIndex == nil {
		m.unexpected("DeleteDocumentIndex")
	}

	return m.deleteDocumentIndex(ctx, documentUUID)
}

func (m *mockRAGUsecase) IndexDocument(ctx context.Context, req usecase.IndexDocumentRequest) (*usecase.IndexDocumentResult, error) {
	if m.indexDocument == nil {
		m.unexpected("IndexDocument")
	}

	return m.indexDocument(ctx, req)
}

func (m *mockRAGUsecase) Query(ctx context.Context, req usecase.QueryRequest) (*usecase.QueryResult, error) {
	if m.query == nil {
		m.unexpected("Query")
	}

	return m.query(ctx, req)
}

func (m *mockRAGUsecase) Ready() bool {
	if m.ready != nil {
		return m.ready()
	}

	return true
}
