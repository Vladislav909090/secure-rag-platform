package grpc

import (
	"context"
	"testing"
	"time"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceDeleteDocumentMapsResult(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, 6, 22, 13, 0, 0, 0, time.UTC)
	uc := NewMockDocumentUsecase(t)
	uc.EXPECT().
		DeleteDocument(mock.Anything, "doc-1").
		RunAndReturn(func(ctx context.Context, docUUID string) (*usecase.DeleteDocumentOutput, error) {
			assert.Equal(t, "doc-1", docUUID)

			return &usecase.DeleteDocumentOutput{DocumentUUID: docUUID, DeletedAt: deletedAt}, nil
		})

	resp, err := (&KnowledgeServiceServerImpl{uc: uc}).DeleteDocument(context.Background(), &pb.DeleteDocumentRequest{
		DocumentUuid: "doc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "doc-1", resp.GetDocumentUuid())
	assert.True(t, resp.GetDeleted())
	assert.Equal(t, deletedAt.Format(time.RFC3339), resp.GetDeletedAt())
}
