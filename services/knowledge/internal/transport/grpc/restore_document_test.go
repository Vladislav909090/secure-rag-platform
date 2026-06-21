package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceRestoreDocumentUsesUsecase(t *testing.T) {
	t.Parallel()

	mock := &mockDocumentUsecase{t: t}
	mock.restoreDocument = func(ctx context.Context, docUUID string) (*model.Document, error) {
		assert.Equal(t, "doc-1", docUUID)

		doc := knowledgeTestDocument(docUUID)
		doc.DeletedAt = nil

		return doc, nil
	}

	resp, err := (&KnowledgeServiceServerImpl{uc: mock}).RestoreDocument(context.Background(), &pb.RestoreDocumentRequest{
		DocumentUuid: "doc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "doc-1", resp.GetDocument().GetUuid())
	assert.Empty(t, resp.GetDocument().GetDeletedAt())
}
