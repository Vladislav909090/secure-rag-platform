package grpc

import (
	"io"
	"strings"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"
)

func (s *KnowledgeServiceServerImpl) CreateDocumentStream(stream pb.KnowledgeService_CreateDocumentStreamServer) error {
	if err := s.requireUC(); err != nil {
		return err
	}

	first, err := stream.Recv()
	if err != nil {
		return toGRPCError(err)
	}

	meta := first.GetMeta()
	if meta == nil || strings.TrimSpace(meta.GetFileName()) == "" {
		return toGRPCError(usecase.ErrInvalidRequest)
	}

	pr, pw := io.Pipe()
	resultCh := make(chan struct {
		out *usecase.CreateDocumentOutput
		err error
	}, 1)

	go func() {
		var attrs map[string]any
		if meta.GetAttributes() != nil {
			attrs = meta.GetAttributes().AsMap()
		}

		var desc *string
		if meta.Description != nil {
			d := meta.GetDescription()
			desc = &d
		}

		out, runErr := s.uc.CreateDocument(stream.Context(), usecase.CreateDocumentInput{
			Title:       meta.GetTitle(),
			Description: desc,
			Attributes:  attrs,
		}, pr, meta.GetFileName(), meta.GetMimeType())
		resultCh <- struct {
			out *usecase.CreateDocumentOutput
			err error
		}{out: out, err: runErr}
	}()

	for {
		msg, recvErr := stream.Recv()
		if recvErr == io.EOF {
			_ = pw.Close()
			break
		}
		if recvErr != nil {
			_ = pw.CloseWithError(recvErr)
			res := <-resultCh
			if res.err != nil {
				return toGRPCError(res.err)
			}

			return toGRPCError(recvErr)
		}

		chunk := msg.GetChunk()
		if chunk == nil || len(chunk.GetData()) == 0 {
			continue
		}

		if _, writeErr := pw.Write(chunk.GetData()); writeErr != nil {
			res := <-resultCh
			if res.err != nil {
				return toGRPCError(res.err)
			}

			return toGRPCError(writeErr)
		}
	}

	res := <-resultCh
	if res.err != nil {
		return toGRPCError(res.err)
	}

	return stream.SendAndClose(&pb.CreateDocumentResponse{
		Document: documentToProto(res.out.Document),
	})
}
