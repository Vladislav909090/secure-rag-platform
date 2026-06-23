package grpc

import (
	"context"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKnowledgeServiceUpdateDocumentAttributesMapsAttributes(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"department": "legal"})
	require.NoError(t, err)

	uc := NewMockDocumentUsecase(t)
	uc.EXPECT().
		UpdateAttributes(mock.Anything, "doc-1", mock.Anything).
		RunAndReturn(func(ctx context.Context, docUUID string, attributes map[string]any) (*model.Document, error) {
			assert.Equal(t, "doc-1", docUUID)
			assert.Equal(t, "legal", attributes["department"])

			doc := knowledgeTestDocument(docUUID)
			doc.Attributes = attributes

			return doc, nil
		})

	resp, err := (&KnowledgeServiceServerImpl{uc: uc}).UpdateDocumentAttributes(context.Background(), &pb.UpdateDocumentAttributesRequest{
		DocumentUuid: "doc-1",
		Attributes:   attrs,
	})
	require.NoError(t, err)
	assert.Equal(t, "legal", resp.GetDocument().GetAttributes().AsMap()["department"])
}
