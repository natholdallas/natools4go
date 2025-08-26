package rands

import (
	"math/rand"
	"time"
)

// FisherYateShuffle shuffle the array, use fisher yate algorithm
func FisherYateShuffle[T any](arr []T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}
