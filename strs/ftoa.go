package strs

import "strconv"

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ParseFloat(f string) float64 {
	result, _ := strconv.ParseFloat(f, 64)
	return result
}
