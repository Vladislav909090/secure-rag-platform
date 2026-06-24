package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryNormalizeRoleCodes(t *testing.T) {
	t.Parallel()

	got := normalizeRoleCodes([]string{" user ", "access_admin", "user", "", "knowledge_editor"})
	want := []string{"access_admin", "knowledge_editor", "user"}
	assert.Equal(t, want, got)

	assert.Nil(t, normalizeRoleCodes(nil))
}
