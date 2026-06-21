package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceGetDocumentUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockDocumentUsecase{t: t}
	mock.getDocument = func(ctx context.Context, docUUID string) (*usecase.DocumentDetail, error) {
		assert.Equal(t, "doc-1", docUUID)

		return &usecase.DocumentDetail{Document: knowledgeTestDocument(docUUID)}, nil
	}

	resp, err := (&KnowledgeServiceServerImpl{uc: mock}).GetDocument(context.Background(), &pb.GetDocumentRequest{
		DocumentUuid: "doc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "doc-1", resp.GetDocument().GetUuid())
	assert.Equal(t, "title", resp.GetDocument().GetTitle())
}
