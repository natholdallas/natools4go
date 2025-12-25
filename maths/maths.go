// Package maths
package maths

// DivCeil performs integer division and rounds the result up (ceiling).
// It handles both positive and negative integers correctly.
func DivCeil(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	} else if (a > 0 && b > 0) || (a < 0 && b < 0) {
		return (a + b - 1) / b
	}
	return (a + b + 1) / b
}

// Digits decomposes an int64 into a slice of its individual digits.
// Example: maths.Digits(123) returns []int64{1, 2, 3}.
func Digits(n int64) []int64 {
	if n == 0 {
		return []int64{0}
	}
	if n < 0 {
		n = -n
	}
	count := 0
	for temp := n; temp != 0; temp /= 10 {
		count++
	}
	digits := make([]int64, count)
	for i := count - 1; i >= 0; i-- {
		digits[i] = n % 10
		n /= 10
	}
	return digits
}
