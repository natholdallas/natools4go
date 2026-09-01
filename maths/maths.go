// Package maths
package maths

import "strconv"

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
	s := strconv.FormatInt(n, 10)
	if len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	digits := make([]int64, len(s))
	for i, c := range s {
		digits[i] = int64(c - '0')
	}
	return digits
}
