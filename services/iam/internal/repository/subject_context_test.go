package repository

import (
	"testing"

	"secure-rag-platform/services/iam/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRolesToCodes(t *testing.T) {
	t.Parallel()

	got := rolesToCodes([]*model.Role{
		{Code: "user"},
		{Code: "access_admin"},
	})
	assert.Equal(t, []string{"user", "access_admin"}, got)
	assert.Empty(t, rolesToCodes(nil))
}
