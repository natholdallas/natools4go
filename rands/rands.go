// Package rands
package rands

import (
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Distribute splits a number into 'parts' random integers that sum up to 'number'.
// It uses the "Dividers" algorithm (Bars and Stars).
func Distribute(total, parts int) []int {
	if parts <= 0 {
		return []int{}
	}
	if parts == 1 {
		return []int{total}
	}
	// Generate random split points
	dividers := make([]int, parts-1)
	for i := 0; i < parts-1; i++ {
		dividers[i] = rand.Intn(total + 1) // +1 allows zero values
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

// Digits extracts 'length' random digits from a large number after shuffling.
func Digits(num *big.Int, length int) (int, error) {
	digits := []byte(num.String())
	FisherYateShuffle(digits)
	return strconv.Atoi(string(digits[:length]))
}

// BetweenTime returns a random time between the start and end range.
func BetweenTime(start, end time.Time) time.Time {
	min := start.UnixNano()
	max := end.UnixNano()
	if min >= max {
		return start
	}
	delta := max - min
	return time.Unix(0, min+rand.Int63n(delta))
}

// Char generates a random alphanumeric string of the specified length.
// It uses pre-allocation and direct byte writing for optimal performance.
func Char(length int) string {
	var b strings.Builder
	b.Grow(length) // Pre-allocate memory to avoid multiple reallocations
	for range length {
		b.WriteByte(alphabet[rand.Intn(len(alphabet))]) // Directly write the byte to avoid string(char) conversion cost
	}
	return b.String()
}

// Pick returns a random element from a slice of any type.
func Pick[T any](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}
	return list[rand.Intn(len(list))]
}
