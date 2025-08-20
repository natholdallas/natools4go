// Package dec provide decimal packaging support
package dec

import (
	"regexp"

	"github.com/natholdallas/natools4go/constraints"
	"github.com/shopspring/decimal"
)

func NewFromInt[T constraints.Integer](value T) decimal.Decimal {
	return decimal.NewFromInt(int64(value))
}

func NewFromString(value string) decimal.Decimal {
	dec, _ := decimal.NewFromString(value)
	return dec
}

func NewFromFloat[T constraints.Float](value T) decimal.Decimal {
	return decimal.NewFromFloat(float64(value))
}

func NewFromFormattedString(value string, replRegexp *regexp.Regexp) decimal.Decimal {
	dec, _ := decimal.NewFromFormattedString(value, replRegexp)
	return dec
}
