package strs

import "strconv"

func FormatUint(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func FormatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}
