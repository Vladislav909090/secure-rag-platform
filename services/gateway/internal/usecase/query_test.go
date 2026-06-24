package usecase

import (
	"context"
	"testing"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	ragv1 "secure-rag-platform/api/gen/go/rag/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGatewayServiceQueryFiltersDocumentsAndUsesDefaults(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleUser)
	expectValidToken(deps, "token", subject)
	deps.knowledge.EXPECT().
		ListDocuments(mock.Anything, mock.Anything).
		Return(&knowledgev1.ListDocumentsResponse{Items: []*knowledgev1.DocumentItem{
			{Document: gatewayDocument("doc-1", map[string]any{"visibility": "public"})},
			{Document: gatewayDocument("doc-2", map[string]any{"visibility": "private"})},
		}}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.MatchedBy(func(attrs map[string]any) bool {
			return attrs["visibility"] == "public"
		})).
		Return(true, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.MatchedBy(func(attrs map[string]any) bool {
			return attrs["visibility"] == "private"
		})).
		Return(false, nil)
	deps.rag.EXPECT().
		Query(mock.Anything, mock.MatchedBy(func(req *ragv1.QueryRequest) bool {
			return req.GetQuery() == "question" &&
				req.GetTopK() == 3 &&
				req.GetEmbeddingModelAlias() == "embed-default" &&
				req.GetGenerationModelAlias() == "gen-default" &&
				len(req.GetDocumentUuids()) == 1 &&
				req.GetDocumentUuids()[0] == "doc-1"
		})).
		Return(&ragv1.QueryResponse{
			Answer:                  "answer",
			ResolvedEmbeddingModel:  "embed-resolved",
			ResolvedGenerationModel: "gen-resolved",
			Contexts: []*ragv1.QueryContext{
				{DocumentUuid: "doc-1", ChunkIndex: 2, Text: "context", Score: 0.8},
			},
		}, nil)

	got, err := svc.Query(context.Background(), QueryRequest{Query: " question "}, "token")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "answer", got.Answer)
	assert.Equal(t, "embed-resolved", got.ResolvedEmbeddingModel)
	require.Len(t, got.Contexts, 1)
	assert.Equal(t, "doc-1", got.Contexts[0].DocumentUUID)
}

func TestGatewayServiceQueryReturnsNoDocumentsFallback(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleUser)
	expectValidToken(deps, "token", subject)
	deps.knowledge.EXPECT().
		ListDocuments(mock.Anything, mock.Anything).
		Return(&knowledgev1.ListDocumentsResponse{Items: []*knowledgev1.DocumentItem{
			{Document: gatewayDocument("doc-1", map[string]any{"visibility": "private"})},
		}}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.Anything).
		Return(false, nil)

	got, err := svc.Query(context.Background(), QueryRequest{Query: "question"}, "token")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "No documents available for this user.", got.Answer)
	assert.Empty(t, got.Contexts)
}

func TestGatewayServiceResolveDocumentsDisableFilter(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	svc.disableFilter = true
	got, err := svc.resolveDocuments(context.Background(), []string{" doc-1 ", "", "doc-2"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"doc-1", "doc-2"}, got)

	deps.knowledge.EXPECT().
		ListDocuments(mock.Anything, mock.Anything).
		Return(&knowledgev1.ListDocumentsResponse{Items: []*knowledgev1.DocumentItem{
			{Document: gatewayDocument("doc-3", map[string]any{})},
			{Document: &knowledgev1.Document{}},
		}}, nil)

	got, err = svc.resolveDocuments(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"doc-3"}, got)
}

func TestGatewayServiceQueryRejectsEmptyQuestion(t *testing.T) {
	t.Parallel()

	svc, _ := newGatewayTestService(t)
	got, err := svc.Query(context.Background(), QueryRequest{Query: " "}, "token")
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, got)
}
