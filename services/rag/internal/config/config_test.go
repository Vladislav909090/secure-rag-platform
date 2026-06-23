package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	t.Setenv(string(Port), "8083")

	assert.Equal(t, "8083", GetValue(Port))
	assert.Empty(t, GetValue(DatabaseDSN))
}
