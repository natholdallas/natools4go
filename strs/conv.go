package strs

import (
	"strconv"

	"github.com/natholdallas/natools4go/constraints"
)

func FormatUint[T constraints.Unsigned](i T) string {
	return strconv.FormatUint(uint64(i), 10)
}

func ParseUint[T constraints.Unsigned](s string) T {
	v, _ := strconv.ParseUint(s, 10, 64)
	return T(v)
}

func FormatInt[T constraints.Signed](i T) string {
	return strconv.FormatInt(int64(i), 10)
}

func ParseInt[T constraints.Signed](s string) T {
	v, _ := strconv.ParseInt(s, 10, 64)
	return T(v)
}

func FormatFloat[T constraints.Float](f T) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func ParseFloat[T constraints.Float](f string) T {
	v, _ := strconv.ParseFloat(f, 64)
	return T(v)
}

func FormatBool(b bool) string {
	return strconv.FormatBool(b)
}

func ParseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}
