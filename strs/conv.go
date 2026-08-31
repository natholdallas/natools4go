package strs

import (
	"strconv"

	"github.com/natholdallas/natools4go/constraints"
)

// uintBits reports the bit width of the unsigned type T.
func uintBits[T constraints.Unsigned]() int {
	switch any(*new(T)).(type) {
	case uint8:
		return 8
	case uint16:
		return 16
	case uint32:
		return 32
	case uint64:
		return 64
	default: // uint, uintptr
		return strconv.IntSize
	}
}

// intBits reports the bit width of the signed type T.
func intBits[T constraints.Signed]() int {
	switch any(*new(T)).(type) {
	case int8:
		return 8
	case int16:
		return 16
	case int32:
		return 32
	case int64:
		return 64
	default: // int
		return strconv.IntSize
	}
}

// floatBits reports the bit width of the float type T.
func floatBits[T constraints.Float]() int {
	switch any(*new(T)).(type) {
	case float32:
		return 32
	default:
		return 64
	}
}

// FormatUint converts an unsigned integer of any type (uint, uint8, uint16, uint32, uint64)
// to its decimal string representation. It casts the input to uint64 internally
// to satisfy the strconv.FormatUint requirement.
func FormatUint[T constraints.Unsigned](i T) string {
	return strconv.FormatUint(uint64(i), 10)
}

// ParseUint parses a decimal string and returns an unsigned integer of the specified type T.
// Parsing is performed with the target type's bit width, so out-of-range values
// (e.g. "300" for uint8) fail and return the zero value instead of silently wrapping.
func ParseUint[T constraints.Unsigned](s string) T {
	v, _ := ParseUintE[T](s)
	return v
}

// ParseUintE is like ParseUint but reports parse and range errors.
func ParseUintE[T constraints.Unsigned](s string) (T, error) {
	v, err := strconv.ParseUint(s, 10, uintBits[T]())
	return T(v), err
}

// FormatInt converts a signed integer of any type (int, int8, int16, int32, int64)
// to its decimal string representation. It casts the input to int64 internally.
func FormatInt[T constraints.Signed](i T) string {
	return strconv.FormatInt(int64(i), 10)
}

// ParseInt parses a decimal string and returns a signed integer of the specified type T.
// Parsing is performed with the target type's bit width, so out-of-range values
// fail and return the zero value instead of silently wrapping.
func ParseInt[T constraints.Signed](s string) T {
	v, _ := ParseIntE[T](s)
	return v
}

// ParseIntE is like ParseInt but reports parse and range errors.
func ParseIntE[T constraints.Signed](s string) (T, error) {
	v, err := strconv.ParseInt(s, 10, intBits[T]())
	return T(v), err
}

// FormatFloat converts a floating-point number (float32 or float64) to a string.
// It uses the 'f' format (no exponent) and the smallest necessary number of
// digits (-1 precision) to represent the value accurately.
func FormatFloat[T constraints.Float](f T) string {
	return strconv.FormatFloat(float64(f), 'f', -1, floatBits[T]())
}

// ParseFloat parses a string into a floating-point number of type T.
// It uses the target type's bit width for the internal parsing. If the string
// is not a valid float, it returns 0.0.
func ParseFloat[T constraints.Float](f string) T {
	v, _ := ParseFloatE[T](f)
	return v
}

// ParseFloatE is like ParseFloat but reports parse and range errors.
func ParseFloatE[T constraints.Float](f string) (T, error) {
	v, err := strconv.ParseFloat(f, floatBits[T]())
	return T(v), err
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
