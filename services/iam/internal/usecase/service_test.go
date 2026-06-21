package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsecaseNormalizeRoleCodes(t *testing.T) {
	t.Parallel()

	got, err := normalizeRoleCodes([]string{" knowledge_editor ", "user", "user", ""})
	require.NoError(t, err)
	want := []string{RoleKnowledgeEditor, RoleUser}
	assert.Equal(t, want, got)

	got, err = normalizeRoleCodes(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{RoleUser}, got)

	_, err = normalizeRoleCodes([]string{"bad"})
	require.ErrorIs(t, err, ErrInvalidArgument)
}

func TestIAMPasswordAndTokenHelpers(t *testing.T) {
	t.Parallel()

	hash, err := hashPassword("secret")
	require.NoError(t, err)
	assert.True(t, checkPassword(hash, "secret"))
	assert.False(t, checkPassword(hash, "other"))
	_, err = hashPassword(" ")
	require.ErrorIs(t, err, ErrInvalidArgument)

	firstTokenHash := hashOpaqueToken("token")
	secondTokenHash := hashOpaqueToken("token")
	assert.Equal(t, firstTokenHash, secondTokenHash)

	token, err := randomToken(8)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRoleHelpers(t *testing.T) {
	t.Parallel()

	roles := []string{RoleUser, RoleAccessAdmin}
	assert.True(t, hasRole(roles, RoleUser))
	assert.True(t, hasAnyRole(roles, RoleSuperAdmin, RoleAccessAdmin))
	assert.False(t, hasAnyRole(roles, RoleSuperAdmin, RoleKnowledgeEditor))
}

func TestMergeAttributes(t *testing.T) {
	t.Parallel()

	got := mergeAttributes(map[string]any{"a": 1, "b": 2}, map[string]any{"b": 3, "c": 4})
	assert.Equal(t, map[string]any{"a": 1, "b": 3, "c": 4}, got)
}
