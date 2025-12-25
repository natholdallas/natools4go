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
	Space  = " "
)

// Wrap ensures the string 'v' starts and ends with 'char'.
// If the prefix or suffix is already present, it will not be duplicated.
func Wrap(v, char string) string {
	return ToEnd(ToStart(v, char), char)
}

// Unwrap removes a single instance of 'char' from both the start and end of 'v'.
// It uses strings.CutPrefix and strings.CutSuffix for precise single-layer removal.
func Unwrap(v, char string) string {
	return TrimEnd(TrimStart(v, char), char)
}

// ToStart prepends 'char' to string 'v' only if 'v' does not already start with 'char'.
func ToStart(v, char string) string {
	if !strings.HasPrefix(v, char) {
		return char + v
	}
	return v
}

// ToEnd appends 'char' to string 'v' only if 'v' does not already end with 'char'.
func ToEnd(v, char string) string {
	if !strings.HasSuffix(v, char) {
		return v + char
	}
	return v
}

// TrimStart removes the first occurrence of the prefix 'char' from 'v'.
// If 'v' does not start with 'char', it returns 'v' unchanged.
func TrimStart(v, char string) string {
	s, _ := strings.CutPrefix(v, char)
	return s
}

// TrimEnd removes the first occurrence of the suffix 'char' from 'v'.
// If 'v' does not end with 'char', it returns 'v' unchanged.
func TrimEnd(v, char string) string {
	s, _ := strings.CutSuffix(v, char)
	return s
}

// AnyPrefix returns true if the string 's' starts with any of the provided 'prefixes'.
func AnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// AnySuffix returns true if the string 's' ends with any of the provided 'suffixes'.
func AnySuffix(s string, suffixes ...string) bool {
	for _, sx := range suffixes {
		if strings.HasSuffix(s, sx) {
			return true
		}
	}
	return false
}
