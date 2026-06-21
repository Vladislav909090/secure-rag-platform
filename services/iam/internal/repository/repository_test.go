package repository

import (
	"encoding/json"
	"testing"
)

func TestAttributesJSONHelpers(t *testing.T) {
	normalized := normalizeAttributes(nil)
	if len(normalized) != 0 {
		t.Fatalf("normalizeAttributes(nil) = %#v, want empty map", normalized)
	}

	raw, err := toJSON(map[string]any{"level": 3, "team": "search"})
	if err != nil {
		t.Fatalf("toJSON() error = %v", err)
	}

	var decoded map[string]any
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded["team"] != "search" || decoded["level"] != float64(3) {
		t.Fatalf("unexpected JSON payload: %#v", decoded)
	}

	attrs, err := fromJSON(raw)
	if err != nil {
		t.Fatalf("fromJSON() error = %v", err)
	}
	if attrs["team"] != "search" || attrs["level"] != float64(3) {
		t.Fatalf("unexpected attrs: %#v", attrs)
	}

	attrs, err = fromJSON(nil)
	if err != nil {
		t.Fatalf("fromJSON(nil) error = %v", err)
	}
	if len(attrs) != 0 {
		t.Fatalf("fromJSON(nil) = %#v, want empty map", attrs)
	}
}

func TestAttributesJSONHelpersRejectUnsupportedValue(t *testing.T) {
	_, err := toJSON(map[string]any{"bad": func() {}})
	if err == nil {
		t.Fatalf("expected toJSON() error for unsupported value")
	}

	_, err = fromJSON([]byte("{"))
	if err == nil {
		t.Fatalf("expected fromJSON() error for malformed JSON")
	}
}
