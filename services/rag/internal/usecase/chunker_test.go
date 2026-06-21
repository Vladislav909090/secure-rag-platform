package usecase

import "testing"

func TestChunkTextTrimsUnicodeAndOverlaps(t *testing.T) {
	chunks := chunkText("  абвгде  ", 3, 1)

	want := []string{"абв", "вгд", "де"}
	if len(chunks) != len(want) {
		t.Fatalf("chunk count = %d, want %d: %v", len(chunks), len(want), chunks)
	}
	for i := range want {
		if chunks[i] != want[i] {
			t.Fatalf("chunk[%d] = %q, want %q", i, chunks[i], want[i])
		}
	}
}

func TestChunkTextDefaultsInvalidOptions(t *testing.T) {
	if got := chunkText("   ", 10, 0); got != nil {
		t.Fatalf("empty text should produce nil chunks, got %v", got)
	}

	chunks := chunkText("abcdef", 3, 3)
	if len(chunks) != 2 || chunks[0] != "abc" || chunks[1] != "def" {
		t.Fatalf("expected overlap >= size to be clamped, got %v", chunks)
	}
}
