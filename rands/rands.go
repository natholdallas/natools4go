// Package rands
package rands

import (
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/natholdallas/natools4go/maths"
)

var (
	chars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	numchars = "aBcDeFgHiJ" // replace 0...9 char
)

// Split random numbers, the number could be zero
func Split(number, parts int) []int {
	if parts <= 0 {
		return []int{}
	}
	if parts == 1 {
		return []int{number}
	}

	// sorter dividers
	dividers := make([]int, parts-1)
	for i := range parts - 1 {
		dividers[i] = rand.Intn(number-1) + 1
	}
	sort.Ints(dividers)

	// collect result
	result := make([]int, parts)
	prev := 0
	for i, d := range dividers {
		result[i] = d - prev
		prev = d
	}
	result[parts-1] = number - prev
	return result
}

// SplitNonZero random split numbers, but filter zero value, result length != parts
// TODO: need optimized
func SplitNonZero(number, parts int) []int {
	numbers := Split(number, parts)
	result := []int{}
	for i := range numbers {
		if numbers[i] == 0 {
			continue
		}
		result = append(result, numbers[i])
	}
	return result
}

// Digits fisher yate algorithm to get bitint's number, must length <= num
func Digits(num *big.Int, length int) (int, error) {
	digits := []byte(num.String())
	FisherYateShuffle(digits)
	return strconv.Atoi(string(digits[:length]))
}

// Time generate random time [start~end]
func Time(start, end time.Time) time.Time {
	min := start.UnixNano()
	max := end.UnixNano()
	nano := min + rand.Int63n(max-min)
	return time.Unix(0, nano)
}

// Deprecated: UniqueChar it will be add more 13 char length
func UniqueChar(length int) string {
	var result strings.Builder
	counter := 0
	for counter < length {
		result.WriteString(string(chars[rand.Intn(len(chars))]))
		counter++
	}
	now := time.Now().UnixMilli()
	for _, i := range maths.SplitDigits(now) {
		result.WriteString(string(numchars[i]))
	}
	return result.String()
}

// Char get random strings
func Char(length int) string {
	var result strings.Builder
	counter := 0
	for counter < length {
		result.WriteString(string(chars[rand.Intn(len(chars))]))
		counter++
	}
	return result.String()
}
