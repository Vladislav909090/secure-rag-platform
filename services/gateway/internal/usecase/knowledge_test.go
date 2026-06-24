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
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestDocumentProtoConversionsAndErrorMapping(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"visibility": "internal"})
	require.NoError(t, err)
	doc := documentFromProto(&knowledgev1.Document{
		Id:             10,
		Uuid:           "doc-1",
		Title:          "policy",
		Description:    "desc",
		Attributes:     attrs,
		FileName:       "policy.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      42,
		ChecksumSha256: "sum",
		StorageKey:     "documents/doc-1/file",
		IndexStatus:    "READY",
		CreatedAt:      "created",
		UpdatedAt:      "updated",
		DeletedAt:      "deleted",
	})
	assert.Equal(t, "doc-1", doc.UUID)
	assert.Equal(t, "internal", doc.Attributes["visibility"])
	assert.Equal(t, int64(42), doc.SizeBytes)

	item := documentItemFromProto(&knowledgev1.DocumentItem{Document: &knowledgev1.Document{Uuid: "doc-2"}})
	assert.Equal(t, "doc-2", item.Document.UUID)

	tests := []struct {
		err  error
		want error
	}{
		{status.Error(codes.NotFound, "missing"), ErrNotFound},
		{status.Error(codes.InvalidArgument, "bad"), ErrInvalidRequest},
		{status.Error(codes.Unauthenticated, "auth"), ErrUnauthorized},
		{status.Error(codes.PermissionDenied, "denied"), ErrForbidden},
	}
	for _, tt := range tests {
		assert.ErrorIs(t, mapUpstreamError(tt.err, "operation"), tt.want)
	}

	unknown := errors.New("upstream exploded")
	require.ErrorIs(t, mapUpstreamError(unknown, "operation"), unknown)
}

func TestGatewayServiceListDocumentsFiltersByPolicy(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleUser)
	expectValidToken(deps, "token", subject)
	allowedDoc := gatewayDocument("doc-allowed", map[string]any{"visibility": "public"})
	deniedDoc := gatewayDocument("doc-denied", map[string]any{"visibility": "private"})
	deps.knowledge.EXPECT().
		ListDocuments(mock.Anything, mock.Anything).
		Return(&knowledgev1.ListDocumentsResponse{Items: []*knowledgev1.DocumentItem{
			{Document: allowedDoc},
			{Document: deniedDoc},
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

	got, err := svc.ListDocuments(context.Background(), "token")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "doc-allowed", got[0].Document.UUID)
}

func TestGatewayServiceGetDocumentDownloadAndDelete(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleKnowledgeEditor)
	expectValidToken(deps, "token", subject)
	expectValidToken(deps, "token", subject)
	doc := gatewayDocument("doc-1", map[string]any{"visibility": "public"})
	deps.knowledge.EXPECT().
		GetDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.GetDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-1"
		})).
		Return(&knowledgev1.GetDocumentResponse{Document: doc}, nil).
		Twice()
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.Anything).
		Return(true, nil).
		Twice()
	deps.knowledge.EXPECT().
		DownloadFile(mock.Anything, mock.MatchedBy(func(req *knowledgev1.DownloadFileRequest) bool {
			return req.GetDocumentUuid() == "doc-1"
		})).
		Return(&httpbody.HttpBody{ContentType: "text/plain", Data: []byte("body")}, nil)
	deps.knowledge.EXPECT().
		DeleteDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.DeleteDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-1"
		})).
		Return(&knowledgev1.DeleteDocumentResponse{DocumentUuid: "doc-1", Deleted: true, DeletedAt: "deleted"}, nil)
	deps.rag.EXPECT().
		DeleteDocumentIndex(mock.Anything, mock.MatchedBy(func(req *ragv1.DeleteDocumentIndexRequest) bool {
			return req.GetDocumentUuid() == "doc-1"
		})).
		Return(&ragv1.DeleteDocumentIndexResponse{DocumentUuid: "doc-1", Deleted: true}, nil)

	file, err := svc.DownloadFile(context.Background(), " doc-1 ", "token")
	require.NoError(t, err)
	assert.Equal(t, "text/plain", file.ContentType)
	assert.Equal(t, []byte("body"), file.Data)

	deleted, err := svc.DeleteDocument(context.Background(), "doc-1", "token")
	require.NoError(t, err)
	assert.True(t, deleted.Deleted)
	assert.True(t, deleted.IndexDeleted)
}

func TestGatewayServiceUpdateDocumentAttributesRejectsBadAttributes(t *testing.T) {
	t.Parallel()

	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleKnowledgeEditor)
	expectValidToken(deps, "token", subject)
	deps.knowledge.EXPECT().
		GetDocument(mock.Anything, mock.Anything).
		Return(&knowledgev1.GetDocumentResponse{Document: gatewayDocument("doc-1", map[string]any{})}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.Anything).
		Return(true, nil)

	got, err := svc.UpdateDocumentAttributes(context.Background(), UpdateDocumentAttributesRequest{
		DocumentUUID: "doc-1",
		Attributes:   map[string]any{"bad": make(chan int)},
	}, "token")
	require.ErrorIs(t, err, ErrInvalidRequest)
	assert.Nil(t, got)
}

func TestGatewayServiceUpdateDocumentUsesAllowedDocument(t *testing.T) {
	t.Parallel()

	title := "updated"
	svc, deps := newGatewayTestService(t)
	subject := gatewaySubject(roleKnowledgeEditor)
	expectValidToken(deps, "token", subject)
	doc := gatewayDocument("doc-1", map[string]any{"visibility": "public"})
	updated := gatewayDocument("doc-1", map[string]any{"visibility": "public"})
	updated.Title = title
	deps.knowledge.EXPECT().
		GetDocument(mock.Anything, mock.Anything).
		Return(&knowledgev1.GetDocumentResponse{Document: doc}, nil)
	deps.policy.EXPECT().
		AllowDocument(mock.Anything, subject, mock.Anything).
		Return(true, nil)
	deps.knowledge.EXPECT().
		UpdateDocument(mock.Anything, mock.MatchedBy(func(req *knowledgev1.UpdateDocumentRequest) bool {
			return req.GetDocumentUuid() == "doc-1" && req.GetTitle() == title
		})).
		Return(&knowledgev1.UpdateDocumentResponse{Document: updated}, nil)

	got, err := svc.UpdateDocument(context.Background(), UpdateDocumentRequest{DocumentUUID: "doc-1", Title: &title}, "token")
	require.NoError(t, err)
	assert.Equal(t, title, got.Document.Title)
}
