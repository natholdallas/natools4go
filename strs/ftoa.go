package strs

import "strconv"

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
