package main

import (
	"context"
	"testing"

	transportgrpc "secure-rag-platform/services/knowledge/internal/transport/grpc"
)

func TestKnowledgeHealthRPC(t *testing.T) {
	server := &transportgrpc.KnowledgeServiceServerImpl{}

	resp, err := server.Health(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if resp.GetStatus() != "ok" {
		t.Fatalf("expected status ok, got %q", resp.GetStatus())
	}
}
