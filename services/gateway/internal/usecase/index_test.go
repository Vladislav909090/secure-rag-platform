package usecase

import (
	"context"
	"errors"
	"testing"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	ragv1 "secure-rag-platform/api/gen/go/rag/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewayServiceReindexDocumentUsesKnowledgeAndRAG(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleKnowledgeEditor)
	expectValidToken(deps, "token", subject)
	deps.knowledge.EXPECT().
		GetDocument(mock.Anything, mock.Anything).
		Return(&knowledgev1.GetDocumentResponse{Document: gatewayDocument("doc-1", map[string]any{"visibility": "public"})}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.Anything).
		Return(true, nil)
	deps.knowledge.EXPECT().
		ReindexDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.ReindexDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-1"
		})).
		Return(&knowledgev1.ReindexDocumentResponse{DocumentUuid: "doc-1", IndexStatus: "PENDING"}, nil)
	deps.rag.EXPECT().
		IndexDocument(mock.Anything, mock.MatchedBy(func(req *ragv1.IndexDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-1" &&
				req.GetEmbeddingModelAlias() == "embed" &&
				req.GetChunkSize() == 128 &&
				req.GetChunkOverlap() == 16
		})).
		Return(&ragv1.IndexDocumentResponse{
			DocumentUuid:           "doc-1",
			ChunkCount:             4,
			EmbeddingDimension:     384,
			ResolvedEmbeddingModel: "embed-resolved",
		}, nil)

	got, err := svc.ReindexDocument(context.Background(), ReindexRequest{
		DocumentUUID:        " doc-1 ",
		EmbeddingModelAlias: "embed",
		ChunkSize:           128,
		ChunkOverlap:        16,
	}, "token")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "doc-1", got.DocumentUUID)
	assert.Equal(t, int32(4), got.ChunkCount)
	assert.Equal(t, "embed-resolved", got.ResolvedEmbeddingModel)
}

func TestGatewayServiceReindexAllDocumentsTracksSuccessAndFailures(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleKnowledgeEditor)
	expectValidToken(deps, "token", subject)
	deps.knowledge.EXPECT().
		ListDocuments(mock.Anything, mock.Anything).
		Return(&knowledgev1.ListDocumentsResponse{Items: []*knowledgev1.DocumentItem{
			{Document: gatewayDocument("doc-ok", map[string]any{"visibility": "public"})},
			{Document: gatewayDocument("doc-fail", map[string]any{"visibility": "public"})},
			{Document: gatewayDocument("doc-denied", map[string]any{"visibility": "private"})},
		}}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.MatchedBy(func(attrs map[string]any) bool {
			return attrs["visibility"] == "public"
		})).
		Return(true, nil).
		Twice()
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.MatchedBy(func(attrs map[string]any) bool {
			return attrs["visibility"] == "private"
		})).
		Return(false, nil)
	deps.knowledge.EXPECT().
		ReindexDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.ReindexDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-ok"
		})).
		Return(&knowledgev1.ReindexDocumentResponse{}, nil)
	deps.rag.EXPECT().
		IndexDocument(mock.Anything, mock.MatchedBy(func(req *ragv1.IndexDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-ok"
		})).
		Return(&ragv1.IndexDocumentResponse{DocumentUuid: "doc-ok", ChunkCount: 3, EmbeddingDimension: 4, ResolvedEmbeddingModel: "embed"}, nil)
	deps.knowledge.EXPECT().
		ReindexDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.ReindexDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-fail"
		})).
		Return((*knowledgev1.ReindexDocumentResponse)(nil), status.Error(codes.NotFound, "missing"))

	got, err := svc.ReindexAllDocuments(context.Background(), ReindexRequest{EmbeddingModelAlias: "embed"}, "token")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int32(2), got.TotalCount)
	assert.Equal(t, int32(1), got.IndexedCount)
	assert.Equal(t, int32(1), got.FailedCount)
	require.Len(t, got.Items, 2)
	assert.True(t, got.Items[0].Indexed)
	assert.Contains(t, got.Items[1].Error, ErrNotFound.Error())
}

func TestGatewayServiceAllowDocumentMapsPolicyErrors(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, mock.Anything, mock.Anything).
		Return(false, errors.New("opa down"))

	allowed, err := svc.allowDocument(context.Background(), gatewaySubject(roleUser), map[string]any{})
	require.ErrorIs(t, err, ErrPolicyUnavailable)
	assert.False(t, allowed)

	svc.policy = nil
	allowed, err = svc.allowDocument(context.Background(), nil, nil)
	require.ErrorIs(t, err, ErrPolicyRequired)
	assert.False(t, allowed)
}
