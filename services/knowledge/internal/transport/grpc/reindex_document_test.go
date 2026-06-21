package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceReindexDocumentMapsResult(t *testing.T) {
	t.Parallel()

	mock := &mockDocumentUsecase{t: t}
	mock.reindexDocument = func(ctx context.Context, docUUID string) (*usecase.ReindexOutput, error) {
		assert.Equal(t, "doc-1", docUUID)

		return &usecase.ReindexOutput{DocumentUUID: docUUID, IndexStatus: model.IndexStatusPending}, nil
	}

	resp, err := (&KnowledgeServiceServerImpl{uc: mock}).ReindexDocument(context.Background(), &pb.ReindexDocumentRequest{
		DocumentUuid: "doc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "doc-1", resp.GetDocumentUuid())
	assert.Equal(t, model.IndexStatusPending, resp.GetIndexStatus())
}
