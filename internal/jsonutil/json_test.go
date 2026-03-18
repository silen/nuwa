package jsonutil

import (
	"testing"
)

func TestJsonStringToAnyDecodesIntoMap(t *testing.T) {
	t.Parallel()

	var out map[string]any
	if err := JsonStringToAny(`{"a":1}`, &out); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if out["a"] == nil {
		t.Fatalf("expected decoded value")
	}
}

func TestAny2AnyCopiesMatchingJSONFields(t *testing.T) {
	t.Parallel()

	type src struct {
		Name string `json:"name"`
	}
	type dst struct {
		Name string `json:"name"`
	}

	var out dst
	if err := Any2Any(src{Name: "nuwa"}, &out); err != nil {
		t.Fatalf("convert error: %v", err)
	}

	if out.Name != "nuwa" {
		t.Fatalf("expected copied name, got %q", out.Name)
	}
}
