// Package random
package random

import (
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/natholdallas/natools4go/maths"
)

var (
	chars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	numchars = "aBcDeFgHiJ" // replace 0...9 char
)

// Name get a random full name, gen by [randomdata.FullName]
func Name() string {
	return randomdata.FullName(randomdata.RandomGender)
}

// Deprecated: Avatar get a random avatar url, gen by [randomdata.SillyName] and website https://avatar.iran.liara.run
func Avatar() string {
	return "https://avatar.iran.liara.run/public/girl?username=" + randomdata.SillyName()
}

// Cover get a random cover preset url, size[400x300], gen by [randomdata.Number] and website https://picsum.photos
func Cover() string {
	return "https://picsum.photos/id/" + strconv.Itoa(randomdata.Number(100)) + "/400/300.webp"
}

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

// UniqueChar it will be add more 13 char length
func UniqueChar(length int) string {
	var result string
	counter := 0
	for counter < length {
		result += string(chars[rand.Intn(len(chars))])
		counter++
	}
	now := time.Now().UnixMilli()
	for _, i := range maths.SplitDigits(now) {
		result += string(numchars[i])
	}
	return result
}

// Char get random strings
func Char(length int) string {
	var result string
	counter := 0
	for counter < length {
		result += string(chars[rand.Intn(len(chars))])
		counter++
	}
	return result
}
