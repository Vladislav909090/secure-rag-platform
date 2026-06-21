package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestKnowledgeToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrDocumentNotFound, codes.NotFound},
		{usecase.ErrFileNotFound, codes.NotFound},
		{usecase.ErrDocumentDeleted, codes.Aborted},
		{usecase.ErrDocumentNotDeleted, codes.Aborted},
		{usecase.ErrFileTooLarge, codes.ResourceExhausted},
		{usecase.ErrInvalidRequest, codes.InvalidArgument},
		{errors.New("upload to storage: down"), codes.Internal},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
