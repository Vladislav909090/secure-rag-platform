package usecase

import "testing"

func TestIsTextMime(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"", true},
		{"text/plain; charset=utf-8", true},
		{"application/json", true},
		{"application/xml", true},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		if got := isTextMime(tt.value); got != tt.want {
			t.Fatalf("isTextMime(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestBuildPrompts(t *testing.T) {
	system, user := buildPrompts("что внутри?", []QueryContext{
		{DocumentUUID: "doc-a", ChunkIndex: 2, Text: "первый факт"},
		{DocumentUUID: "doc-b", ChunkIndex: 1, Text: "второй факт"},
	})

	if system == "" {
		t.Fatalf("expected non-empty system prompt")
	}
	if user != "Вопрос:\nчто внутри?\n\nКонтекст:\n[doc-a:2]\nпервый факт\n\n[doc-b:1]\nвторой факт" {
		t.Fatalf("unexpected user prompt:\n%s", user)
	}
}
