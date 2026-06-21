package grpc

import (
	"context"
	"errors"
	"testing"

	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestExtractBearerToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", " Bearer token "))
	token, err := extractBearerToken(ctx)
	if err != nil {
		t.Fatalf("extractBearerToken() error = %v", err)
	}
	if token != "token" {
		t.Fatalf("extractBearerToken() = %q", token)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "token"))
	if _, err := extractBearerToken(ctx); !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for malformed header, got %v", err)
	}
}

func TestAccessHelpers(t *testing.T) {
	principal := &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleUser}}
	if !canAccessUser(principal, "u1") {
		t.Fatalf("user should access self")
	}
	if canAccessUser(principal, "u2") {
		t.Fatalf("plain user should not access another user")
	}

	admin := &usecase.Principal{UserID: "u1", Roles: []string{usecase.RoleAccessAdmin}}
	if !canAccessUser(admin, "u2") || !isAdmin(admin) {
		t.Fatalf("access admin should access another user")
	}
}

func TestToGRPCError(t *testing.T) {
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
