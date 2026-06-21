package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
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
