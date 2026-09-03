// Package rands
package rands

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"sort"
	"time"
)

var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Distribute splits a number into 'parts' random integers that sum up to 'number'.
// It uses the "Dividers" algorithm (Bars and Stars).
func Distribute(total, parts int) []int {
	if parts <= 0 || total < 0 {
		return []int{}
	}
	if parts == 1 {
		return []int{total}
	}
	// Generate random split points
	dividers := make([]int, parts-1)
	for i := 0; i < parts-1; i++ {
		dividers[i] = rand.IntN(total + 1) // +1 allows zero values
	}
	sort.Ints(dividers)

	result := make([]int, parts)
	prev := 0
	for i, d := range dividers {
		result[i] = d - prev
		prev = d
	}
	result[parts-1] = total - prev
	return result
}

// DistributeStrict splits a number into 'parts' random integers where each part is >= 1.
// It ensures the resulting slice length is always equal to 'parts'.
func DistributeStrict(total, parts int) []int {
	if parts <= 0 || total < parts {
		return []int{}
	}

	// Implementation trick: distribute (total - parts), then add 1 to each part.
	// This ensures no zero values while maintaining the sum.
	v := Distribute(total-parts, parts)
	for i := range v {
		v[i]++
	}
	return v
}

// Digits extracts 'length' random digits from a large number, returned as a
// string of exactly 'length' characters. Leading zeros are preserved (unlike
// parsing to an int). The sign (if any) is discarded first so '-' cannot land
// in the middle of the result. A partial Fisher-Yates shuffle samples exactly
// 'length' distinct positions, running in O(length).
func Digits(num *big.Int, length int) (string, error) {
	digits := []byte(num.String())
	if len(digits) > 0 && digits[0] == '-' {
		digits = digits[1:]
	}
	if length < 1 || length > len(digits) {
		return "", fmt.Errorf("length %d out of range [1, %d]", length, len(digits))
	}
	// Partial Fisher-Yates: place a uniformly random distinct digit at each of
	// the first 'length' positions, then take those.
	for i := range length {
		j := rand.IntN(len(digits)-i) + i
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits[:length]), nil
}

// BetweenTime returns a random time between the start and end range.
func BetweenTime(start, end time.Time) time.Time {
	min := start.UnixNano()
	max := end.UnixNano()
	if min >= max {
		return start
	}
	delta := max - min
	return time.Unix(0, min+rand.Int64N(delta))
}

// Char generates a random alphanumeric string of the specified length.
// It uses pre-allocation and direct byte writing for optimal performance.
func Char(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[rand.IntN(len(alphabet))]
	}
	return string(b)
}

// Pick returns a random element from a slice of any type.
func Pick[T any](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}
	return list[rand.IntN(len(list))]
}
