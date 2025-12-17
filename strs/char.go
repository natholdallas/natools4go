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

func EnsureAroundChar(v, char string) string {
	v = EnsureStartChar(v, char)
	v = EnsureEndChar(v, char)
	return v
}

func EnsureNoAroundChar(v, char string) string {
	v = EnsureNoStartChar(v, char)
	v = EnsureNoEndChar(v, char)
	return v
}

func EnsureStartChar(v, char string) string {
	if !strings.HasPrefix(v, char) {
		v = char + v
	}
	return v
}

func EnsureEndChar(v, char string) string {
	if !strings.HasSuffix(v, char) {
		v = v + char
	}
	return v
}

func EnsureNoStartChar(v, char string) string {
	s, _ := strings.CutPrefix(v, char)
	return s
}

func EnsureNoEndChar(v, char string) string {
	s, _ := strings.CutSuffix(v, char)
	return s
}

func HasPrefix(s string, prefixes ...string) bool {
	for _, i := range prefixes {
		if strings.HasPrefix(s, i) {
			return true
		}
	}
	return false
}

func HasSuffix(s string, prefixes ...string) bool {
	for _, i := range prefixes {
		if strings.HasSuffix(s, i) {
			return true
		}
	}
	return false
}
