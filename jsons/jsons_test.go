package jsons

import "testing"

func TestGetOK(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{"b": "x"},
	}
	v, ok := GetOK(m, "a", "b")
	if !ok || v != "x" {
		t.Errorf("GetOK = %v,%v want x,true", v, ok)
	}
	if _, ok := GetOK(m, "a", "missing"); ok {
		t.Error("GetOK missing should be false")
	}
	if _, ok := GetOK(m, "nope", "b"); ok {
		t.Error("GetOK missing parent should be false")
	}
}

func TestGetNilSafe(t *testing.T) {
	m := map[string]any{}
	if got := Get(m, "x", "y"); got != nil {
		t.Errorf("Get should return nil, got %v", got)
	}
}

func TestSet(t *testing.T) {
	m := map[string]any{"a": map[string]any{}}
	Set(m, 42, "a", "k")
	if m["a"].(map[string]any)["k"] != 42 {
		t.Error("Set did not assign value")
	}
	// Missing intermediate -> no-op, must not panic.
	Set(m, 1, "missing", "k")
}

func TestMarshalString(t *testing.T) {
	if s, err := String(map[string]int{"a": 1}); err != nil || s != `{"a":1}` {
		t.Errorf("String = %q,%v", s, err)
	}
}
