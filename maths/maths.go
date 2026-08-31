// Package maths
package maths

// DivCeil performs integer division and rounds the result up (ceiling).
// It handles both positive and negative integers correctly. A zero divisor
// returns 0.
func DivCeil(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	q, r := a/b, a%b
	if r != 0 && (a >= 0) == (b >= 0) {
		q++
	}
	return q
}

// Digits decomposes an int64 into a slice of its individual digits,
// discarding the sign. Example: maths.Digits(123) returns []int64{1, 2, 3}.
func Digits(n int64) []int64 {
	if n == 0 {
		return []int64{0}
	}
	var u uint64
	if n < 0 {
		// Avoid overflow on MinInt64 by negating in unsigned space.
		u = uint64(-(n + 1)) + 1
	} else {
		u = uint64(n)
	}
	count := 0
	for temp := u; temp != 0; temp /= 10 {
		count++
	}
	digits := make([]int64, count)
	for i := count - 1; i >= 0; i-- {
		digits[i] = int64(u % 10)
		u /= 10
	}
	return digits
}
