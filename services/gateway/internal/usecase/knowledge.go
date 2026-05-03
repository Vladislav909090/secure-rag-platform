package usecase

import (
	"context"
	"fmt"
	"strings"

	iamv1 "secure-rag-platform/services/iam/gen/v1"
	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) ListDocuments(ctx context.Context, accessToken string) ([]DocumentWithVersions, error) {
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

	out := make([]DocumentWithVersions, 0, len(resp.GetItems()))
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
		out = append(out, documentWithVersionsFromProto(item))
	}

	return out, nil
}

func (s *Service) GetDocument(
	ctx context.Context,
	documentUUID string,
	accessToken string,
) (*DocumentWithVersions, error) {
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

	return &DocumentWithVersions{
		Document: documentFromProto(resp.GetDocument()),
		Versions: versionsFromProto(resp.GetVersions()),
	}, nil
}

func (s *Service) GetDocumentVersion(
	ctx context.Context,
	documentUUID string,
	versionNumber int32,
	accessToken string,
) (*DocumentWithVersions, error) {
	if !s.Ready() {
		return nil, ErrNotConfigured
	}

	documentUUID = strings.TrimSpace(documentUUID)
	if documentUUID == "" || versionNumber <= 0 {
		return nil, ErrInvalidRequest
	}

	subject, err := s.validateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.knowledge.GetDocumentVersion(ctx, &knowledgev1.GetDocumentVersionRequest{
		DocumentUuid:  documentUUID,
		VersionNumber: versionNumber,
	})
	if err != nil {
		return nil, mapUpstreamError(err, "get document version")
	}

	allowed, err := s.isDocumentAllowed(ctx, subject, resp.GetDocument())
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	return &DocumentWithVersions{
		Document: documentFromProto(resp.GetDocument()),
		Versions: []DocumentVersion{
			versionFromProto(resp.GetVersion()),
		},
	}, nil
}

func (s *Service) DownloadFile(
	ctx context.Context,
	documentUUID string,
	versionNumber int32,
	accessToken string,
) (*FileResult, error) {
	if versionNumber > 0 {
		if _, err := s.GetDocumentVersion(ctx, documentUUID, versionNumber, accessToken); err != nil {
			return nil, err
		}

		resp, err := s.knowledge.DownloadVersionFile(ctx, &knowledgev1.DownloadVersionFileRequest{
			DocumentUuid:  strings.TrimSpace(documentUUID),
			VersionNumber: versionNumber,
		})
		if err != nil {
			return nil, mapUpstreamError(err, "download version file")
		}
		return &FileResult{ContentType: resp.GetContentType(), Data: resp.GetData()}, nil
	}

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

func documentWithVersionsFromProto(item *knowledgev1.DocumentWithVersions) DocumentWithVersions {
	if item == nil {
		return DocumentWithVersions{}
	}
	return DocumentWithVersions{
		Document: documentFromProto(item.GetDocument()),
		Versions: versionsFromProto(item.GetVersions()),
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
		ID:                   doc.GetId(),
		UUID:                 doc.GetUuid(),
		Title:                doc.GetTitle(),
		Description:          doc.GetDescription(),
		Attributes:           attrs,
		CurrentVersionNumber: doc.GetCurrentVersionNumber(),
		CreatedAt:            doc.GetCreatedAt(),
		UpdatedAt:            doc.GetUpdatedAt(),
		DeletedAt:            doc.GetDeletedAt(),
	}
}

func versionsFromProto(versions []*knowledgev1.DocumentVersion) []DocumentVersion {
	out := make([]DocumentVersion, 0, len(versions))
	for _, version := range versions {
		out = append(out, versionFromProto(version))
	}
	return out
}

func versionFromProto(version *knowledgev1.DocumentVersion) DocumentVersion {
	if version == nil {
		return DocumentVersion{}
	}
	return DocumentVersion{
		ID:             version.GetId(),
		UUID:           version.GetUuid(),
		DocumentID:     version.GetDocumentId(),
		VersionNumber:  version.GetVersionNumber(),
		FileName:       version.GetFileName(),
		FileExtension:  version.GetFileExtension(),
		MimeType:       version.GetMimeType(),
		SizeBytes:      version.GetSizeBytes(),
		ChecksumSHA256: version.GetChecksumSha256(),
		StorageKey:     version.GetStorageKey(),
		IndexStatus:    version.GetIndexStatus(),
		CreatedAt:      version.GetCreatedAt(),
	}
}
