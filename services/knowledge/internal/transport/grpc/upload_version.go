package grpc

import (
	"io"
	"strings"

	pb "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"
)

func (s *KnowledgeServiceServerImpl) UploadVersionStream(stream pb.KnowledgeService_UploadVersionStreamServer) error {
	if err := s.requireUC(); err != nil {
		return err
	}

	first, err := stream.Recv()
	if err != nil {
		return toGRPCError(err)
	}

	meta := first.GetMeta()
	if meta == nil || strings.TrimSpace(meta.GetDocumentUuid()) == "" || strings.TrimSpace(meta.GetFileName()) == "" {
		return toGRPCError(usecase.ErrInvalidRequest)
	}

	pr, pw := io.Pipe()
	resultCh := make(chan struct {
		out *usecase.UploadVersionOutput
		err error
	}, 1)

	go func() {
		out, runErr := s.uc.UploadVersion(
			stream.Context(),
			meta.GetDocumentUuid(),
			pr,
			meta.GetFileName(),
			meta.GetMimeType(),
		)
		resultCh <- struct {
			out *usecase.UploadVersionOutput
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

	return stream.SendAndClose(&pb.UploadVersionResponse{
		DocumentId:           res.out.DocumentID,
		DocumentUuid:         res.out.DocumentUUID,
		CurrentVersionNumber: res.out.CurrentVersion,
		Version:              versionToProto(res.out.Version),
	})
}
