package grpc

import (
	"context"
	"io"
	"testing"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type testCreateDocumentStream struct {
	grpc.ServerStream

	ctx      context.Context
	requests []*pb.CreateDocumentStreamRequest
	sent     *pb.CreateDocumentResponse
}

func (s *testCreateDocumentStream) Context() context.Context {
	return s.ctx
}

func (s *testCreateDocumentStream) Recv() (*pb.CreateDocumentStreamRequest, error) {
	if len(s.requests) == 0 {
		return nil, io.EOF
	}

	req := s.requests[0]
	s.requests = s.requests[1:]

	return req, nil
}

func (s *testCreateDocumentStream) SendAndClose(resp *pb.CreateDocumentResponse) error {
	s.sent = resp

	return nil
}

func TestKnowledgeServiceCreateDocumentStreamUploadsChunks(t *testing.T) {
	t.Parallel()

	attrs, err := structpb.NewStruct(map[string]any{"department": "search"})
	require.NoError(t, err)

	desc := "file description"
	mock := &mockDocumentUsecase{
		t: t,
		createDocument: func(_ context.Context, input usecase.CreateDocumentInput, file io.Reader, fileName string, mimeType string) (*usecase.CreateDocumentOutput, error) {
			data, readErr := io.ReadAll(file)
			require.NoError(t, readErr)

			assert.Equal(t, "knowledge note", input.Title)
			require.NotNil(t, input.Description)
			assert.Equal(t, desc, *input.Description)
			assert.Equal(t, map[string]any{"department": "search"}, input.Attributes)
			assert.Equal(t, "note.txt", fileName)
			assert.Equal(t, "text/plain", mimeType)
			assert.Equal(t, "hello world", string(data))

			return &usecase.CreateDocumentOutput{Document: knowledgeTestDocument("doc-1")}, nil
		},
	}
	stream := &testCreateDocumentStream{
		ctx: context.Background(),
		requests: []*pb.CreateDocumentStreamRequest{
			{
				Payload: &pb.CreateDocumentStreamRequest_Meta{
					Meta: &pb.CreateDocumentMeta{
						Title:       "knowledge note",
						Description: &desc,
						Attributes:  attrs,
						FileName:    "note.txt",
						MimeType:    "text/plain",
					},
				},
			},
			{
				Payload: &pb.CreateDocumentStreamRequest_Chunk{
					Chunk: &pb.FileChunk{Data: []byte("hello ")},
				},
			},
			{
				Payload: &pb.CreateDocumentStreamRequest_Chunk{
					Chunk: &pb.FileChunk{},
				},
			},
			{
				Payload: &pb.CreateDocumentStreamRequest_Chunk{
					Chunk: &pb.FileChunk{Data: []byte("world")},
				},
			},
		},
	}

	err = (&KnowledgeServiceServerImpl{uc: mock}).CreateDocumentStream(stream)
	require.NoError(t, err)
	require.NotNil(t, stream.sent)
	require.NotNil(t, stream.sent.GetDocument())
	assert.Equal(t, "doc-1", stream.sent.GetDocument().GetUuid())
}

func TestKnowledgeServiceCreateDocumentStreamRejectsMissingFileName(t *testing.T) {
	t.Parallel()

	stream := &testCreateDocumentStream{
		ctx: context.Background(),
		requests: []*pb.CreateDocumentStreamRequest{
			{
				Payload: &pb.CreateDocumentStreamRequest_Meta{
					Meta: &pb.CreateDocumentMeta{Title: "knowledge note"},
				},
			},
		},
	}

	err := (&KnowledgeServiceServerImpl{uc: &mockDocumentUsecase{t: t}}).CreateDocumentStream(stream)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Nil(t, stream.sent)
}
