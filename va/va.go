// Package va is tiny packaging support validator
package va

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// va is the singleton validator instance used across the package.
var va = validator.New()

// Struct validates a struct's exported fields based on their 'validate' tags.
// If validation fails, it returns a formatted error string containing
// the field name, actual value, and the failed validation tag/parameter.
func Struct(data any) error {
	errs := va.Struct(data)
	if errs == nil {
		return nil
	}

	vErrs, ok := errs.(validator.ValidationErrors)
	if !ok {
		return errs
	}

	results := make([]string, 0, len(vErrs))
	for _, e := range vErrs {
		var b strings.Builder
		param := e.Param()
		field := e.Field()
		value := e.Value()
		tag := e.Tag()
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

// Var validates a single variable against a specific validation tag.
// Example: Var("admin@example.com", "required,email")
func Var(s any, tag string) error {
	return va.Var(s, tag)
}
