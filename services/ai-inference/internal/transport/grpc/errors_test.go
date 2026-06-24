package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/ai-inference/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAIInferenceToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrInvalidArgument, codes.InvalidArgument},
		{usecase.ErrAliasNotFound, codes.NotFound},
		{usecase.ErrAliasTaskMismatch, codes.FailedPrecondition},
		{usecase.ErrProviderNotConfigured, codes.Unavailable},
		{usecase.ErrProviderFailed, codes.Unavailable},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
