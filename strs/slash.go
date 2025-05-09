package strs

import "strings"

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
