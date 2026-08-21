package constraints

import (
	"testing"
)

// Compile-time assertions that the constraint interfaces accept the intended
// built-in types and reject nothing at the type-parameter level.
func integer[T Integer](v T) T { return v }
func signed[T Signed](v T) T   { return v }
func unsigned[T Unsigned](v T) T { return v }
func floating[T Float](v T) T  { return v }

func TestIntegerAcceptsBuiltins(t *testing.T) {
	cases := []any{
		integer(int(1)),
		integer(int8(1)),
		integer(int16(1)),
		integer(int32(1)),
		integer(int64(1)),
		integer(uint(1)),
		integer(uint8(1)),
		integer(uint16(1)),
		integer(uint32(1)),
		integer(uint64(1)),
		integer(uintptr(1)),
	}
	if len(cases) != 11 {
		t.Fatalf("expected 11 accepted types, got %d", len(cases))
	}
}

func TestSignedUnsignedFloat(t *testing.T) {
	_ = signed(int(-1))
	_ = unsigned(uint(1))
	_ = floating(float32(1.5))
	_ = floating(float64(2.5))
}