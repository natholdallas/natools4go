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

// Some used to check has any data in arr
func Some() {}

// Filter used to filter arr's data
func Filter() {}

// ForEach used to for loop arrs
func ForEach() {}
