package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/rag/internal/usecase"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRAGToGRPCError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrNotConfigured, codes.Unavailable},
		{usecase.ErrInvalidRequest, codes.InvalidArgument},
		{usecase.ErrNoContexts, codes.NotFound},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.code, status.Code(toGRPCError(tt.err)))
		})
	}
}
