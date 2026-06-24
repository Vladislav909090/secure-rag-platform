package grpc

import "testing"

func TestGatewayMapToStructFallback(t *testing.T) {
	out := mapToStruct(map[string]any{"bad": make(chan int)})
	if out == nil || len(out.GetFields()) != 0 {
		t.Fatalf("invalid struct input should produce empty struct, got %#v", out)
	}

	if out = mapToStruct(nil); out == nil {
		t.Fatalf("nil map should produce an empty struct")
	}
}
