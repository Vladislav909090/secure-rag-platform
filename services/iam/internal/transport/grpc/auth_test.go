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

func TestIAMExtractBearerToken(t *testing.T) {
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

func TestIAMAccessHelpers(t *testing.T) {
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

func TestIAMRequireUC(t *testing.T) {
	if status.Code(requireUC(nil)) != codes.Unavailable {
		t.Fatalf("requireUC(nil) should return unavailable")
	}
	if err := requireUC(usecase.NewIAMUsecase(nil, nil, usecase.DefaultConfig(), nil)); err != nil {
		t.Fatalf("requireUC(non-nil) error = %v", err)
	}
}
