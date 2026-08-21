package strs

import "testing"

func TestWrap(t *testing.T) {
	if got := Wrap("a", "/"); got != "/a/" {
		t.Errorf("Wrap = %q, want /a/", got)
	}
	if got := Wrap("/a/", "/"); got != "/a/" {
		t.Errorf("Wrap existing = %q, want /a/", got)
	}
}

func TestUnwrap(t *testing.T) {
	if got := Unwrap("/a/", "/"); got != "a" {
		t.Errorf("Unwrap = %q, want a", got)
	}
}

func TestToStartToEnd(t *testing.T) {
	if got := ToStart("a", "/"); got != "/a" {
		t.Errorf("ToStart = %q, want /a", got)
	}
	if got := ToEnd("a", "/"); got != "a/" {
		t.Errorf("ToEnd = %q, want a/", got)
	}
}

func TestTrimStartTrimEnd(t *testing.T) {
	if got := TrimStart("/a", "/"); got != "a" {
		t.Errorf("TrimStart = %q, want a", got)
	}
	if got := TrimEnd("a/", "/"); got != "a" {
		t.Errorf("TrimEnd = %q, want a", got)
	}
}

func TestAnyPrefixSuffix(t *testing.T) {
	if !AnyPrefix("hello", "he", "xx") {
		t.Error("AnyPrefix should be true")
	}
	if AnyPrefix("hello", "xx") {
		t.Error("AnyPrefix should be false")
	}
	if !AnySuffix("hello", "lo") {
		t.Error("AnySuffix should be true")
	}
}
