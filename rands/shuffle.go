package rands

import (
	"math/rand"
)

// FisherYateShuffle shuffle the array, use fisher yate algorithm
func FisherYateShuffle[T any](arr []T) {
	if len(arr) <= 1 {
		return
	}
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}
