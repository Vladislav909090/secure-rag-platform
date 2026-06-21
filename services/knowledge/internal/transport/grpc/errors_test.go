package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/knowledge/internal/usecase"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestKnowledgeToGRPCError(t *testing.T) {
	t.Parallel()

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
		tt := tt
		t.Run(tt.code.String(), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.code, status.Code(toGRPCError(tt.err)))
		})
	}
}
