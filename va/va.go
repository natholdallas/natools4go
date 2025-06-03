package va

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var va = validator.New()

func Struct(data any) error {
	errs := va.Struct(data)
	if errs == nil {
		return nil
	}
	results := []string{}
	for _, err := range errs.(validator.ValidationErrors) {
		var b strings.Builder
		param := err.Param()
		field := err.Field()
		value := err.Value()
		tag := err.Tag()
		b.WriteString("[")
		b.WriteString(field)
		if v := fmt.Sprint(value); v != "" {
			b.WriteString(":")
			b.WriteString(v)
		}
		b.WriteString("]:[")
		if param == "" {
			b.WriteString(tag)
			b.WriteString("]")
		} else {
			b.WriteString(tag)
			b.WriteString("-")
			b.WriteString(param)
			b.WriteString("]")
		}
		results = append(results, b.String())
	}
	return errors.New(strings.Join(results, "\n"))
}

func Var(s any, tag string) error {
	return va.Var(s, tag)
}
