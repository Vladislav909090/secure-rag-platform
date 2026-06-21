package usecase

import (
	"errors"
	"testing"
)

func TestNormalizeRoleCodes(t *testing.T) {
	got, err := normalizeRoleCodes([]string{" knowledge_editor ", "user", "user", ""})
	if err != nil {
		t.Fatalf("normalizeRoleCodes() error = %v", err)
	}
	want := []string{RoleKnowledgeEditor, RoleUser}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("normalizeRoleCodes() = %v, want %v", got, want)
	}

	got, err = normalizeRoleCodes(nil)
	if err != nil {
		t.Fatalf("normalizeRoleCodes(nil) error = %v", err)
	}
	if len(got) != 1 || got[0] != RoleUser {
		t.Fatalf("empty roles should default to user, got %v", got)
	}

	if _, err = normalizeRoleCodes([]string{"bad"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestPasswordAndTokenHelpers(t *testing.T) {
	hash, err := hashPassword("secret")
	if err != nil {
		t.Fatalf("hashPassword() error = %v", err)
	}
	if !checkPassword(hash, "secret") {
		t.Fatalf("expected password check to pass")
	}
	if checkPassword(hash, "other") {
		t.Fatalf("expected password check to fail")
	}
	_, err = hashPassword(" ")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}

	firstTokenHash := hashOpaqueToken("token")
	secondTokenHash := hashOpaqueToken("token")
	if firstTokenHash != secondTokenHash {
		t.Fatalf("hashOpaqueToken should be deterministic")
	}

	token, err := randomToken(8)
	if err != nil {
		t.Fatalf("randomToken() error = %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestRoleHelpers(t *testing.T) {
	roles := []string{RoleUser, RoleAccessAdmin}
	if !hasRole(roles, RoleUser) {
		t.Fatalf("expected user role")
	}
	if !hasAnyRole(roles, RoleSuperAdmin, RoleAccessAdmin) {
		t.Fatalf("expected access admin role")
	}
	if hasAnyRole(roles, RoleSuperAdmin, RoleKnowledgeEditor) {
		t.Fatalf("did not expect super/admin editor role")
	}
}

func TestMergeAttributes(t *testing.T) {
	got := mergeAttributes(map[string]any{"a": 1, "b": 2}, map[string]any{"b": 3, "c": 4})
	if got["a"] != 1 || got["b"] != 3 || got["c"] != 4 {
		t.Fatalf("unexpected merged attrs: %#v", got)
	}
}
