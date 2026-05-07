// Package structs is tiny packaging support structs
package structs

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/natholdallas/natools4go/va"
)

func Map(input any) map[string]any {
	return To[map[string]any](input)
}

func To[T any](input any) T {
	s := new(T)
	mapstructure.Decode(input, s)
	return *s
}

func Vo[T any](input any) (T, error) {
	s := new(T)
	mapstructure.Decode(input, s)
	return *s, va.Struct(*s)
}
