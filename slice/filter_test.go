package slice

import (
	"encoding/json"
	"testing"
)

// TestFilterResultIsNeverNil guards the JSON contract: a nil slice marshals to
// null, which breaks consumers expecting an array ("content: null").
func TestFilterResultIsNeverNil(t *testing.T) {
	got := Filter([]int{1, 2, 3}, func(int) bool { return false })
	if got == nil {
		t.Fatal("Filter returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Fatalf("Filter returned %v, want empty", got)
	}

	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "[]" {
		t.Errorf("JSON = %s, want []", b)
	}
}

func TestFilterKeepsMatches(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4}, func(i int) bool { return i%2 == 0 })
	if len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Errorf("Filter = %v, want [2 4]", got)
	}
}

// TestFilterNilInput covers the nil-input edge case, which must also yield a
// non-nil result so handlers never emit null.
func TestFilterNilInput(t *testing.T) {
	var in []string
	got := Filter(in, func(string) bool { return true })
	if got == nil {
		t.Fatal("Filter(nil) returned nil, want an empty slice")
	}
}
