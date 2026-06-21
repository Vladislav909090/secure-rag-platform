package grpc

import (
	"testing"
	"time"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/usecase"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestIAMProtoConversions(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 30, 0, 0, time.UTC)
	revokedAt := now.Add(time.Hour)

	subject := subjectToProto(&model.SubjectContext{
		UserID:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      []string{usecase.RoleUser},
		Attributes: map[string]any{"department": "search"},
		CtxVer:     7,
	})
	if subject.GetUserId() != "u1" || subject.GetAttributes().AsMap()["department"] != "search" {
		t.Fatalf("unexpected subject proto: %#v", subject)
	}

	user := userToProto(&model.UserView{
		ID:         "u1",
		Login:      "alice",
		IsActive:   true,
		CtxVer:     7,
		Roles:      []string{usecase.RoleUser},
		Attributes: map[string]any{"level": 2},
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Minute),
	})
	if user.GetId() != "u1" || user.GetCreatedAt() != now.Format(time.RFC3339) {
		t.Fatalf("unexpected user proto: %#v", user)
	}

	role := roleToProto(&model.Role{
		ID:          1,
		Code:        usecase.RoleUser,
		Name:        "User",
		Description: "Default user",
		CreatedAt:   now,
	})
	if role.GetCode() != usecase.RoleUser || role.GetCreatedAt() != now.Format(time.RFC3339) {
		t.Fatalf("unexpected role proto: %#v", role)
	}

	roles := rolesToProto([]*model.Role{{ID: 1, Code: usecase.RoleUser, CreatedAt: now}, nil})
	if len(roles) != 2 || roles[0].GetCode() != usecase.RoleUser || roles[1] != nil {
		t.Fatalf("unexpected roles proto: %#v", roles)
	}

	session := sessionToProto(&model.UserSession{
		ID:        "s1",
		UserID:    "u1",
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: &revokedAt,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
	})
	if session.GetId() != "s1" || session.GetRevokedAt() != revokedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected session proto: %#v", session)
	}
}

func TestIAMProtoConversionsNilAndFallbacks(t *testing.T) {
	if subjectToProto(nil) != nil {
		t.Fatalf("subjectToProto(nil) should return nil")
	}
	if userToProto(nil) != nil {
		t.Fatalf("userToProto(nil) should return nil")
	}
	if roleToProto(nil) != nil {
		t.Fatalf("roleToProto(nil) should return nil")
	}
	if sessionToProto(nil) != nil {
		t.Fatalf("sessionToProto(nil) should return nil")
	}

	empty := mapToStruct(nil)
	if len(empty.GetFields()) != 0 {
		t.Fatalf("mapToStruct(nil) = %#v, want empty struct", empty)
	}

	fallback := mapToStruct(map[string]any{"bad": func() {}})
	if len(fallback.GetFields()) != 0 {
		t.Fatalf("mapToStruct(unsupported) = %#v, want empty struct", fallback)
	}

	if got := structToMap(nil); len(got) != 0 {
		t.Fatalf("structToMap(nil) = %#v, want empty map", got)
	}

	value, err := structpb.NewStruct(map[string]any{"a": "b"})
	if err != nil {
		t.Fatalf("structpb.NewStruct() error = %v", err)
	}
	if got := structToMap(value); got["a"] != "b" {
		t.Fatalf("structToMap() = %#v", got)
	}
}
