package strs

import "strconv"

func FormatUint(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func ParseUint(s string) uint64 {
	result, _ := strconv.ParseUint(s, 10, 64)
	return result
}

func FormatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ParseInt(s string) int64 {
	result, _ := strconv.ParseInt(s, 10, 64)
	return result
}
