package grpc

import (
	"errors"
	"testing"

	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewayToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrNotConfigured, codes.Unavailable},
		{usecase.ErrInvalidRequest, codes.InvalidArgument},
		{usecase.ErrUnauthorized, codes.Unauthenticated},
		{usecase.ErrForbidden, codes.PermissionDenied},
		{usecase.ErrNotFound, codes.NotFound},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
