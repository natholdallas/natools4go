package random

import (
	"math/rand"
	"time"
)

// shuffle the array, use fisher yate algorithm
func FisherYateShuffle[T any](arr []T) []T {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}
