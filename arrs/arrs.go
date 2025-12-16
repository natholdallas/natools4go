// Package arrs
package arrs

// Map used to convert arr's content to other type
func Map[T, R any](arr []T, f func(T) R) []R {
	res := make([]R, len(arr))
	for i, v := range arr {
		res[i] = f(v)
	}
	return res
}

// Filter used to filter arr's data
func Filter[T any](s []T, f func(T) bool) []T {
	r := make([]T, 0)
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

// ForEach used to for loop arrs
func ForEach[T any](s []T, f func(T)) {
	for _, v := range s {
		f(v)
	}
}
