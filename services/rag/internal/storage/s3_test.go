package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Storage(t *testing.T) {
	t.Parallel()

	got, err := NewS3Storage("127.0.0.1:9000", "access", "secret", "bucket", false)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "bucket", got.bucket)
	assert.NotNil(t, got.client)
}

func TestNewS3StorageRejectsInvalidEndpoint(t *testing.T) {
	t.Parallel()

	got, err := NewS3Storage("://bad", "access", "secret", "bucket", false)
	require.Error(t, err)
	assert.Nil(t, got)
}
