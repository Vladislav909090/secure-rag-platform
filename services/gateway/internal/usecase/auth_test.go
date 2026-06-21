package usecase

import (
	"context"
	"errors"
	"testing"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAccessRoleHelpers(t *testing.T) {
	if err := requireAdmin(nil); err != nil {
		t.Fatalf("requireAdmin(nil) error = %v", err)
	}
	if err := requireAdmin(&iamv1.SubjectContext{Roles: []string{roleAccessAdmin}}); err != nil {
		t.Fatalf("requireAdmin(access_admin) error = %v", err)
	}
	if err := requireAdmin(&iamv1.SubjectContext{Roles: []string{roleUser}}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("requireAdmin(user) error = %v, want forbidden", err)
	}

	if err := requireDocumentEditor(nil); err != nil {
		t.Fatalf("requireDocumentEditor(nil) error = %v", err)
	}
	if err := requireDocumentEditor(&iamv1.SubjectContext{Roles: []string{roleKnowledgeEditor}}); err != nil {
		t.Fatalf("requireDocumentEditor(editor) error = %v", err)
	}
	if err := requireDocumentEditor(&iamv1.SubjectContext{Roles: []string{roleAccessAdmin}}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("requireDocumentEditor(access_admin) error = %v, want forbidden", err)
	}
}

func TestOutgoingAuthContext(t *testing.T) {
	ctx, err := outgoingAuthContext(context.Background(), " token ")
	if err != nil {
		t.Fatalf("outgoingAuthContext() error = %v", err)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatalf("expected outgoing metadata")
	}
	if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer token" {
		t.Fatalf("unexpected authorization metadata: %v", got)
	}

	if _, err = outgoingAuthContext(context.Background(), " "); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for empty token, got %v", err)
	}
}

func TestAuthProtoConversions(t *testing.T) {
	tokens := tokenPairFromProto(&iamv1.TokenPairResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresIn:    30,
		TokenType:    "Bearer",
	})
	if tokens.AccessToken != "access" || tokens.RefreshToken != "refresh" || tokens.ExpiresIn != 30 || tokens.TokenType != "Bearer" {
		t.Fatalf("unexpected token pair: %#v", tokens)
	}
	if empty := tokenPairFromProto(nil); empty.AccessToken != "" || empty.RefreshToken != "" {
		t.Fatalf("nil token response should produce empty token pair, got %#v", empty)
	}

	attrs, err := structpb.NewStruct(map[string]any{"department": "legal"})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	subject := subjectFromProto(&iamv1.SubjectContext{
		UserId:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      []string{roleUser},
		Attributes: attrs,
		CtxVer:     7,
	})
	if subject.UserID != "u1" || subject.Login != "alice" || subject.Attributes["department"] != "legal" || subject.CtxVer != 7 {
		t.Fatalf("unexpected subject: %#v", subject)
	}
	if subjectFromProto(nil) != nil {
		t.Fatalf("nil subject should remain nil")
	}
}
