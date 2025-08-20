// Package strs
package strs

import (
	"strings"
)

const (
	Dot    = "."
	Slash  = "/"
	Comma  = ","
	Strike = "-"
)

func EnsureAroundSlash(v string) string {
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	if !strings.HasSuffix(v, "/") {
		v = v + "/"
	}
	return v
}

func EnsureStartSlash(v string) string {
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	return v
}

func EnsureEndSlash(v string) string {
	if !strings.HasSuffix(v, "/") {
		v = v + "/"
	}
	return v
}

func EnsureNoStartSlash(v string) string {
	s, _ := strings.CutPrefix(v, "/")
	return s
}

func EnsureNoEndSlash(v string) string {
	s, _ := strings.CutSuffix(v, "/")
	return s
}
