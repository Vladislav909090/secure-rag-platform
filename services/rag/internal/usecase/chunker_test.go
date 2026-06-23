package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkTextTrimsUnicodeAndOverlaps(t *testing.T) {
	t.Parallel()

	chunks := chunkText("  абвгде  ", 3, 1)

	want := []string{"абв", "вгд", "де"}
	assert.Equal(t, want, chunks)
}

func TestChunkTextDefaultsInvalidOptions(t *testing.T) {
	t.Parallel()

	assert.Nil(t, chunkText("   ", 10, 0))

	chunks := chunkText("abcdef", 3, 3)
	assert.Equal(t, []string{"abc", "def"}, chunks)
}
