package rands

import (
	"math/rand/v2"
)

// FisherYatesShuffle shuffles the slice in place using the Fisher-Yates
// algorithm. Empty and single-element slices are no-ops (handled by rand.Shuffle).
func FisherYatesShuffle[T any](arr []T) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}
