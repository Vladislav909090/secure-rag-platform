package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	t.Setenv(string(Port), "8080")

	assert.Equal(t, "8080", GetValue(Port))
}

func TestGetFirstValue(t *testing.T) {
	t.Setenv(string(LegacyDBDSN), "legacy")
	t.Setenv(string(DatabaseDSN), "primary")

	assert.Equal(t, "primary", GetFirstValue(DatabaseDSN, LegacyDBDSN))
	assert.Equal(t, "legacy", GetFirstValue(RedisPassword, LegacyDBDSN))
	assert.Empty(t, GetFirstValue(RedisPassword))
}
