package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceListDocumentsMapsItems(t *testing.T) {
	t.Parallel()

	mock := &mockDocumentUsecase{t: t}
	mock.listDocuments = func(ctx context.Context) ([]*usecase.DocumentDetail, error) {
		return []*usecase.DocumentDetail{
			{Document: knowledgeTestDocument("doc-1")},
			{Document: knowledgeTestDocument("doc-2")},
		}, nil
	}

	resp, err := (&KnowledgeServiceServerImpl{uc: mock}).ListDocuments(context.Background(), &pb.ListDocumentsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetItems(), 2)
	assert.Equal(t, "doc-1", resp.GetItems()[0].GetDocument().GetUuid())
	assert.Equal(t, "doc-2", resp.GetItems()[1].GetDocument().GetUuid())
}
