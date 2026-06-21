package grpc

import (
	"context"
	"errors"
	"testing"

	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestExtractAccessToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", " Bearer token "))
	if got := extractAccessToken(ctx); got != "token" {
		t.Fatalf("extractAccessToken() = %q", got)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "raw-token"))
	if got := extractAccessToken(ctx); got != "raw-token" {
		t.Fatalf("extractAccessToken() raw = %q", got)
	}
}

func TestToGRPCError(t *testing.T) {
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
