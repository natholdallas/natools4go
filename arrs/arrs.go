// Package arrs
package arrs

// Map applies the function f to each element of the slice arr and returns
// a new slice containing the results. It transforms a slice of type T
// into a slice of type R.
func Map[T, R any](arr []T, f func(T) R) []R {
	res := make([]R, len(arr))
	for i, v := range arr {
		res[i] = f(v)
	}
	return res
}

// Filter iterates over elements of the slice s, returning a new slice
// containing all elements for which the predicate f returns true.
func Filter[T any](s []T, f func(T) bool) []T {
	r := make([]T, 0, len(s))
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

// ForEach executes the provided function f once for each element
// in the slice s. It is typically used for side effects.
func ForEach[T any](s []T, f func(T)) {
	for _, v := range s {
		f(v)
	}
}

// GetDefault returns the first element of the provided variadic arguments (args)
// if available; otherwise, it returns the specified defaultValue.
//
// This is a generic helper used to simulate optional parameters with
// default values in Go functions.
func GetDefault[T any](defaultValue T, args []T) T {
	if len(args) > 0 {
		return args[0]
	}
	return defaultValue
}
