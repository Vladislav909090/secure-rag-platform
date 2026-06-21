package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIAMToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrInvalidArgument, codes.InvalidArgument},
		{usecase.ErrRateLimited, codes.ResourceExhausted},
		{usecase.ErrForbidden, codes.PermissionDenied},
		{usecase.ErrNotFound, codes.NotFound},
		{usecase.ErrUserExists, codes.AlreadyExists},
		{usecase.ErrInactiveUser, codes.FailedPrecondition},
		{usecase.ErrUnauthorized, codes.Unauthenticated},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
