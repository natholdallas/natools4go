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

// get a random full name, gen by [randomdata.FullName]
func Name() string {
	return randomdata.FullName(randomdata.RandomGender)
}

// get a random avatar url, gen by [randomdata.SillyName] and website https://avatar.iran.liara.run
func Avatar() string {
	return "https://avatar.iran.liara.run/public/girl?username=" + randomdata.SillyName()
}

// get a random cover preset url, size[400x300], gen by [randomdata.Number] and website https://picsum.photos
func Cover() string {
	return "https://picsum.photos/id/" + strconv.Itoa(randomdata.Number(40, 200)) + "/400/300.webp"
}

// random split numbers, the number could be zero
func Split(number int, parts int) []int {
	if parts <= 0 {
		return []int{}
	}
	if parts == 1 {
		return []int{number}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dividers := make([]int, parts-1)
	for i := range parts - 1 {
		dividers[i] = r.Intn(number-1) + 1
	}
	sort.Ints(dividers)

	result := make([]int, parts)
	prev := 0
	for i, d := range dividers {
		result[i] = d - prev
		prev = d
	}
	result[parts-1] = number - prev
	return result
}

// fisher yate algorithm to get bitint's number, must length <= num
func Digits(num *big.Int, length int) (int, error) {
	digits := []byte(num.String())
	FisherYateShuffle(digits)
	return strconv.Atoi(string(digits[:length]))
}

// generate random time [start~end]
func Time(start, end time.Time) time.Time {
	min := start.UnixNano()
	max := end.UnixNano()
	randomNano := min + rand.Int63n(max-min)
	return time.Unix(0, randomNano)
}

// it will be add more 13 char length
func UniqueStr(length int) string {
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

// get random strings
func Str(length int) string {
	var result string
	counter := 0
	for counter < length {
		result += string(chars[rand.Intn(len(chars))])
		counter++
	}
	return result
}
