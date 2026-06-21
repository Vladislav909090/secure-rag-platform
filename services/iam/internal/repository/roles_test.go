package repository

import (
	"reflect"
	"testing"
)

func TestRepositoryNormalizeRoleCodes(t *testing.T) {
	got := normalizeRoleCodes([]string{" user ", "access_admin", "user", "", "knowledge_editor"})
	want := []string{"access_admin", "knowledge_editor", "user"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeRoleCodes() = %v, want %v", got, want)
	}

	if got = normalizeRoleCodes(nil); got != nil {
		t.Fatalf("normalizeRoleCodes(nil) = %v, want nil", got)
	}
}
