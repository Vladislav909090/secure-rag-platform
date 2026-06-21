package grpc

import (
	"testing"

	"secure-rag-platform/services/gateway/internal/usecase"
)

func TestAuthProxyProtoConversions(t *testing.T) {
	tokens := tokenPairToProto(&usecase.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresIn:    60,
		TokenType:    "Bearer",
	})
	if tokens.GetAccessToken() != "access" || tokens.GetRefreshToken() != "refresh" || tokens.GetExpiresIn() != 60 {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
	if empty := tokenPairToProto(nil); empty.GetAccessToken() != "" {
		t.Fatalf("nil tokens should produce empty response, got %#v", empty)
	}

	subject := subjectToProto(&usecase.SubjectContext{
		UserID:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      []string{"user"},
		Attributes: map[string]any{"team": "platform"},
		CtxVer:     4,
	})
	if subject.GetUserId() != "u1" || subject.GetAttributes().AsMap()["team"] != "platform" {
		t.Fatalf("unexpected subject: %#v", subject)
	}
	if subjectToProto(nil) != nil {
		t.Fatalf("nil subject should remain nil")
	}
}
