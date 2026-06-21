package grpc

import (
	"context"
	"io"

	"secure-rag-platform/services/knowledge/internal/model"
	"secure-rag-platform/services/knowledge/internal/usecase"
)

type documentUsecase interface {
	CreateDocument(ctx context.Context, input usecase.CreateDocumentInput, file io.Reader, fileName string, mimeType string) (*usecase.CreateDocumentOutput, error)
	DeleteDocument(ctx context.Context, docUUID string) (*usecase.DeleteDocumentOutput, error)
	DownloadFile(ctx context.Context, docUUID string) (*usecase.FileDownload, error)
	GetDocument(ctx context.Context, docUUID string) (*usecase.DocumentDetail, error)
	ListDocuments(ctx context.Context) ([]*usecase.DocumentDetail, error)
	ReindexDocument(ctx context.Context, docUUID string) (*usecase.ReindexOutput, error)
	RestoreDocument(ctx context.Context, docUUID string) (*model.Document, error)
	UpdateAttributes(ctx context.Context, docUUID string, attributes map[string]any) (*model.Document, error)
	UpdateDocument(ctx context.Context, docUUID string, title *string, description *string) (*model.Document, error)
}

func usecaseOrNil(uc *usecase.DocumentUsecase) documentUsecase {
	if uc == nil {
		return nil
	}

	return uc
}
