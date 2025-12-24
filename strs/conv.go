package strs

import (
	"strconv"

	"github.com/natholdallas/natools4go/constraints"
)

// FormatUint converts an unsigned integer of any type (uint, uint8, uint16, uint32, uint64)
// to its decimal string representation. It casts the input to uint64 internally
// to satisfy the strconv.FormatUint requirement.
func FormatUint[T constraints.Unsigned](i T) string {
	return strconv.FormatUint(uint64(i), 10)
}

// ParseUint parses a decimal string and returns an unsigned integer of the specified type T.
// Note: If parsing fails, it returns the zero value of type T. It assumes a 64-bit
// range during parsing before casting to the target type.
func ParseUint[T constraints.Unsigned](s string) T {
	v, _ := strconv.ParseUint(s, 10, 64)
	return T(v)
}

// FormatInt converts a signed integer of any type (int, int8, int16, int32, int64)
// to its decimal string representation. It casts the input to int64 internally.
func FormatInt[T constraints.Signed](i T) string {
	return strconv.FormatInt(int64(i), 10)
}

// ParseInt parses a decimal string and returns a signed integer of the specified type T.
// Note: If the string is invalid, it returns the zero value (0). The parsing is
// performed with a 64-bit precision limit.
func ParseInt[T constraints.Signed](s string) T {
	v, _ := strconv.ParseInt(s, 10, 64)
	return T(v)
}

// FormatFloat converts a floating-point number (float32 or float64) to a string.
// It uses the 'f' format (no exponent) and the smallest necessary number of
// digits (-1 precision) to represent the value accurately.
func FormatFloat[T constraints.Float](f T) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

// ParseFloat parses a string into a floating-point number of type T.
// It uses 64-bit precision for the internal parsing process. If the string
// is not a valid float, it returns 0.0.
func ParseFloat[T constraints.Float](f string) T {
	v, _ := strconv.ParseFloat(f, 64)
	return T(v)
}

// FormatBool returns "true" or "false" based on the boolean value of b.
// It is a direct wrapper around the standard library's [strconv.FormatBool].
func FormatBool(b bool) string {
	return strconv.FormatBool(b)
}

// ParseBool interprets a string and returns the boolean value it represents.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Any other value returns false.
func ParseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}

