package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceUpdateDocumentMapsOptionalFields(t *testing.T) {
	t.Parallel()

	title := "new title"
	description := "new description"
	mock := &mockDocumentUsecase{t: t}
	mock.updateDocument = func(ctx context.Context, docUUID string, gotTitle *string, gotDescription *string) (*model.Document, error) {
		assert.Equal(t, "doc-1", docUUID)
		require.NotNil(t, gotTitle)
		assert.Equal(t, title, *gotTitle)
		require.NotNil(t, gotDescription)
		assert.Equal(t, description, *gotDescription)

		doc := knowledgeTestDocument(docUUID)
		doc.Title = *gotTitle
		doc.Description = gotDescription

		return doc, nil
	}

	resp, err := (&KnowledgeServiceServerImpl{uc: mock}).UpdateDocument(context.Background(), &pb.UpdateDocumentRequest{
		DocumentUuid: "doc-1",
		Title:        &title,
		Description:  &description,
	})
	require.NoError(t, err)
	assert.Equal(t, title, resp.GetDocument().GetTitle())
	assert.Equal(t, description, resp.GetDocument().GetDescription())
}
