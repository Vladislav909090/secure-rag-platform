package grpc

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/require"
)

type mockDocumentUsecase struct {
	t *testing.T

	createDocument   func(context.Context, usecase.CreateDocumentInput, io.Reader, string, string) (*usecase.CreateDocumentOutput, error)
	deleteDocument   func(context.Context, string) (*usecase.DeleteDocumentOutput, error)
	downloadFile     func(context.Context, string) (*usecase.FileDownload, error)
	getDocument      func(context.Context, string) (*usecase.DocumentDetail, error)
	listDocuments    func(context.Context) ([]*usecase.DocumentDetail, error)
	reindexDocument  func(context.Context, string) (*usecase.ReindexOutput, error)
	restoreDocument  func(context.Context, string) (*model.Document, error)
	updateAttributes func(context.Context, string, map[string]any) (*model.Document, error)
	updateDocument   func(context.Context, string, *string, *string) (*model.Document, error)
}

var _ documentUsecase = (*mockDocumentUsecase)(nil)

func (m *mockDocumentUsecase) unexpected(name string) {
	if m.t != nil {
		m.t.Helper()
		require.FailNowf(m.t, "unexpected usecase call", "unexpected usecase call: %s", name)
	}
	panic(fmt.Sprintf("unexpected usecase call: %s", name))
}

func (m *mockDocumentUsecase) CreateDocument(ctx context.Context, input usecase.CreateDocumentInput, file io.Reader, fileName string, mimeType string) (*usecase.CreateDocumentOutput, error) {
	if m.createDocument == nil {
		m.unexpected("CreateDocument")
	}

	return m.createDocument(ctx, input, file, fileName, mimeType)
}

func (m *mockDocumentUsecase) DeleteDocument(ctx context.Context, docUUID string) (*usecase.DeleteDocumentOutput, error) {
	if m.deleteDocument == nil {
		m.unexpected("DeleteDocument")
	}

	return m.deleteDocument(ctx, docUUID)
}

func (m *mockDocumentUsecase) DownloadFile(ctx context.Context, docUUID string) (*usecase.FileDownload, error) {
	if m.downloadFile == nil {
		m.unexpected("DownloadFile")
	}

	return m.downloadFile(ctx, docUUID)
}

func (m *mockDocumentUsecase) GetDocument(ctx context.Context, docUUID string) (*usecase.DocumentDetail, error) {
	if m.getDocument == nil {
		m.unexpected("GetDocument")
	}

	return m.getDocument(ctx, docUUID)
}

func (m *mockDocumentUsecase) ListDocuments(ctx context.Context) ([]*usecase.DocumentDetail, error) {
	if m.listDocuments == nil {
		m.unexpected("ListDocuments")
	}

	return m.listDocuments(ctx)
}

func (m *mockDocumentUsecase) ReindexDocument(ctx context.Context, docUUID string) (*usecase.ReindexOutput, error) {
	if m.reindexDocument == nil {
		m.unexpected("ReindexDocument")
	}

	return m.reindexDocument(ctx, docUUID)
}

func (m *mockDocumentUsecase) RestoreDocument(ctx context.Context, docUUID string) (*model.Document, error) {
	if m.restoreDocument == nil {
		m.unexpected("RestoreDocument")
	}

	return m.restoreDocument(ctx, docUUID)
}

func (m *mockDocumentUsecase) UpdateAttributes(ctx context.Context, docUUID string, attributes map[string]any) (*model.Document, error) {
	if m.updateAttributes == nil {
		m.unexpected("UpdateAttributes")
	}

	return m.updateAttributes(ctx, docUUID, attributes)
}

func (m *mockDocumentUsecase) UpdateDocument(ctx context.Context, docUUID string, title *string, description *string) (*model.Document, error) {
	if m.updateDocument == nil {
		m.unexpected("UpdateDocument")
	}

	return m.updateDocument(ctx, docUUID, title, description)
}

func knowledgeTestDocument(uuid string) *model.Document {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	desc := "description"

	return &model.Document{
		ID:             1,
		UUID:           uuid,
		Title:          "title",
		Description:    &desc,
		Attributes:     map[string]any{"department": "search"},
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      9,
		ChecksumSHA256: "checksum",
		StorageKey:     "documents/" + uuid + "/file",
		IndexStatus:    model.IndexStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
