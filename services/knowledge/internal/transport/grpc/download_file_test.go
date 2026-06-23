package grpc

import (
	"context"
	"io"
	"strings"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeServiceDownloadFileReadsBody(t *testing.T) {
	t.Parallel()

	uc := NewMockDocumentUsecase(t)
	uc.EXPECT().
		DownloadFile(mock.Anything, "doc-1").
		RunAndReturn(func(ctx context.Context, docUUID string) (*usecase.FileDownload, error) {
			assert.Equal(t, "doc-1", docUUID)

			return &usecase.FileDownload{
				Body:     io.NopCloser(strings.NewReader("file body")),
				FileName: "file.txt",
				MimeType: "text/plain",
			}, nil
		})

	resp, err := (&KnowledgeServiceServerImpl{uc: uc}).DownloadFile(context.Background(), &pb.DownloadFileRequest{
		DocumentUuid: "doc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "text/plain", resp.GetContentType())
	assert.Equal(t, []byte("file body"), resp.GetData())
}
