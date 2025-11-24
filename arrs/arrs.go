// Package arrs
package arrs

func Map[T, R any](arr []T, f func(T) R) []R {
	res := make([]R, len(arr))
	for i, v := range arr {
		res[i] = f(v)
	}
	return res
}
