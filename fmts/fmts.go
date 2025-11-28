// Package fmts is tiny packaging support fmt
package fmts

import "fmt"

func Read[T any](field string) T {
	var v T
	fmt.Print(field + ": ")
	fmt.Scanln(&v)
	return v
}
