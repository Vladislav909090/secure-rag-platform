package usecase

import (
	"context"
	"fmt"
	"strings"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	ragv1 "secure-rag-platform/api/gen/go/rag/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Service) ListDocuments(ctx context.Context, accessToken string) ([]DocumentItem, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.knowledge.ListDocuments(ctx, &knowledgev1.ListDocumentsRequest{})
	if err != nil {
		return nil, mapUpstreamError(err, "list documents")
	}

	out := make([]DocumentItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		doc := item.GetDocument()
		if doc == nil {
			continue
		}
		allowed, err := s.isDocumentAllowed(ctx, subject, doc)
		if err != nil {
			return nil, err
		}
		if !allowed {
			continue
		}
		out = append(out, documentItemFromProto(item))
	}

	return out, nil
}

func (s *Service) GetDocument(
	ctx context.Context,
	documentUUID string,
	accessToken string,
) (*DocumentItem, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	documentUUID = strings.TrimSpace(documentUUID)
	if documentUUID == "" {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.knowledge.GetDocument(ctx, &knowledgev1.GetDocumentRequest{DocumentUuid: documentUUID})
	if err != nil {
		return nil, mapUpstreamError(err, "get document")
	}

	allowed, err := s.isDocumentAllowed(ctx, subject, resp.GetDocument())
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	return &DocumentItem{
		Document: documentFromProto(resp.GetDocument()),
	}, nil
}

func (s *Service) DownloadFile(
	ctx context.Context,
	documentUUID string,
	accessToken string,
) (*FileResult, error) {
	if _, err := s.GetDocument(ctx, documentUUID, accessToken); err != nil {
		return nil, err
	}

	resp, err := s.knowledge.DownloadFile(ctx, &knowledgev1.DownloadFileRequest{
		DocumentUuid: strings.TrimSpace(documentUUID),
	})
	if err != nil {
		return nil, mapUpstreamError(err, "download file")
	}

	return &FileResult{ContentType: resp.GetContentType(), Data: resp.GetData()}, nil
}

func (s *Service) DeleteDocument(
	ctx context.Context,
	documentUUID string,
	accessToken string,
) (*DeleteDocumentResult, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	documentUUID = strings.TrimSpace(documentUUID)
	if documentUUID == "" {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	doc, err := s.getAllowedDocument(ctx, documentUUID, subject)
	if err != nil {
		return nil, err
	}
	documentUUID = doc.GetUuid()

	resp, err := s.knowledge.DeleteDocument(ctx, &knowledgev1.DeleteDocumentRequest{
		DocumentUuid: documentUUID,
	})
	if err != nil {
		return nil, mapUpstreamError(err, "delete document")
	}

	if _, err := s.rag.DeleteDocumentIndex(ctx, &ragv1.DeleteDocumentIndexRequest{DocumentUuid: documentUUID}); err != nil {
		return nil, fmt.Errorf("delete document index: %w", err)
	}

	return &DeleteDocumentResult{
		DocumentUUID: resp.GetDocumentUuid(),
		Deleted:      resp.GetDeleted(),
		DeletedAt:    resp.GetDeletedAt(),
		IndexDeleted: true,
	}, nil
}

// UpdateDocument обновляет основные поля документа через gateway
func (s *Service) UpdateDocument(
	ctx context.Context,
	req UpdateDocumentRequest,
	accessToken string,
) (*DocumentItem, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	documentUUID := strings.TrimSpace(req.DocumentUUID)
	if documentUUID == "" || (req.Title == nil && req.Description == nil) {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	doc, err := s.getAllowedDocument(ctx, documentUUID, subject)
	if err != nil {
		return nil, err
	}

	resp, err := s.knowledge.UpdateDocument(ctx, &knowledgev1.UpdateDocumentRequest{
		DocumentUuid: doc.GetUuid(),
		Title:        req.Title,
		Description:  req.Description,
	})
	if err != nil {
		return nil, mapUpstreamError(err, "update document")
	}

	return &DocumentItem{Document: documentFromProto(resp.GetDocument())}, nil
}

// UpdateDocumentAttributes заменяет атрибуты документа через gateway
func (s *Service) UpdateDocumentAttributes(
	ctx context.Context,
	req UpdateDocumentAttributesRequest,
	accessToken string,
) (*DocumentItem, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	documentUUID := strings.TrimSpace(req.DocumentUUID)
	if documentUUID == "" {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	doc, err := s.getAllowedDocument(ctx, documentUUID, subject)
	if err != nil {
		return nil, err
	}

	attrs, err := structpb.NewStruct(req.Attributes)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	resp, err := s.knowledge.UpdateDocumentAttributes(ctx, &knowledgev1.UpdateDocumentAttributesRequest{
		DocumentUuid: doc.GetUuid(),
		Attributes:   attrs,
	})
	if err != nil {
		return nil, mapUpstreamError(err, "update document attributes")
	}

	return &DocumentItem{Document: documentFromProto(resp.GetDocument())}, nil
}

func (s *Service) OpenCreateDocumentStream(
	ctx context.Context,
	accessToken string,
) (knowledgev1.KnowledgeService_CreateDocumentStreamClient, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if err = requireDocumentEditor(subject); err != nil {
		return nil, err
	}

	stream, err := s.knowledge.CreateDocumentStream(ctx)
	if err != nil {
		return nil, mapUpstreamError(err, "create document stream")
	}

	return stream, nil
}

func (s *Service) getAllowedDocument(
	ctx context.Context,
	documentUUID string,
	subject *iamv1.SubjectContext,
) (*knowledgev1.Document, error) {
	resp, err := s.knowledge.GetDocument(ctx, &knowledgev1.GetDocumentRequest{DocumentUuid: documentUUID})
	if err != nil {
		return nil, mapUpstreamError(err, "get document")
	}

	doc := resp.GetDocument()
	allowed, err := s.isDocumentAllowed(ctx, subject, doc)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	return doc, nil
}

func (s *Service) isDocumentAllowed(
	ctx context.Context,
	subject *iamv1.SubjectContext,
	doc *knowledgev1.Document,
) (bool, error) {
	if doc == nil {
		return false, ErrNotFound
	}
	attrs := map[string]any{}
	if doc.GetAttributes() != nil {
		attrs = doc.GetAttributes().AsMap()
	}

	return s.allowDocument(ctx, subject, attrs)
}

func mapUpstreamError(err error, operation string) error {
	switch status.Code(err) {
	case codes.NotFound:
		return ErrNotFound
	case codes.InvalidArgument:
		return ErrInvalidRequest
	case codes.Unauthenticated:
		return ErrUnauthorized
	case codes.PermissionDenied:
		return ErrForbidden
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func documentItemFromProto(item *knowledgev1.DocumentItem) DocumentItem {
	if item == nil {
		return DocumentItem{}
	}

	return DocumentItem{
		Document: documentFromProto(item.GetDocument()),
	}
}

func documentFromProto(doc *knowledgev1.Document) Document {
	if doc == nil {
		return Document{}
	}
	attrs := map[string]any{}
	if doc.GetAttributes() != nil {
		attrs = doc.GetAttributes().AsMap()
	}

	return Document{
		ID:             doc.GetId(),
		UUID:           doc.GetUuid(),
		Title:          doc.GetTitle(),
		Description:    doc.GetDescription(),
		Attributes:     attrs,
		FileName:       doc.GetFileName(),
		FileExtension:  doc.GetFileExtension(),
		MimeType:       doc.GetMimeType(),
		SizeBytes:      doc.GetSizeBytes(),
		ChecksumSHA256: doc.GetChecksumSha256(),
		StorageKey:     doc.GetStorageKey(),
		IndexStatus:    doc.GetIndexStatus(),
		CreatedAt:      doc.GetCreatedAt(),
		UpdatedAt:      doc.GetUpdatedAt(),
		DeletedAt:      doc.GetDeletedAt(),
	}
}
