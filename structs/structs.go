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
	v, _ := ToE[T](input)
	return v
}

// ToE decodes input into a new T, returning the mapstructure error (if any).
func ToE[T any](input any) (T, error) {
	s := new(T)
	err := mapstructure.Decode(input, s)
	return *s, err
}

func Vo[T any](input any) (T, error) {
	v, err := ToE[T](input)
	if err != nil {
		return v, err
	}
	return v, va.Struct(v)
}
