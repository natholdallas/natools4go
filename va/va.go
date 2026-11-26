// Package va is tiny packaging support validator
package va

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the singleton validator instance used across the package.
var validate = validator.New()

// sensitiveFragments are matched as case-insensitive substrings of the struct
// field name. They cover the names that actually carry credentials while
// avoiding over-broad tokens: a bare "key" would redact SortKey/Mapping and a
// bare "pin" would redact Mapping, so both are deliberately absent.
var sensitiveFragments = []string{
	"password", "passwd", "passphrase", "secret", "token",
	"apikey", "api_key", "privatekey", "private_key",
	"accesskey", "access_key", "salt", "credential",
	"authorization", "cookie", "connstring", "connection_string", "dsn",
}

// isSensitive reports whether a field name carries a credential whose value
// must never be echoed into an error message. Validation errors are routinely
// returned verbatim to HTTP clients (see fext.Fail handlers), so echoing the
// offending value would leak plaintext passwords, secrets and tokens.
func isSensitive(field string) bool {
	l := strings.ToLower(field)
	for _, p := range sensitiveFragments {
		if strings.Contains(l, p) {
			return true
		}
	}
	// Suffix/other forms that must not match in the middle of a word: a
	// Passport field is not a password, while Pwd always is.
	return strings.HasSuffix(l, "pwd") ||
		strings.HasSuffix(l, "hash") ||
		strings.Contains(l, "otp") ||
		strings.Contains(l, "cvv")
}

// redact returns a placeholder for sensitive field names and the original value
// otherwise.
func redact(field, value string) string {
	if value == "" {
		return ""
	}
	if isSensitive(field) {
		return "***"
	}
	return value
}

// FieldError describes a single failed validation without exposing the value.
type FieldError struct {
	Field string // struct field name
	Tag   string // validation tag that failed (e.g. "required", "min")
	Param string // tag parameter (e.g. "4" for min=4)
}

// String renders the error in the historical "[Field:value]:[tag-param]" shape
// with sensitive values redacted.
func (e FieldError) String() string {
	return fmt.Sprintf("[%s]:[%s]", e.Field, joinTag(e.Tag, e.Param))
}

func joinTag(tag, param string) string {
	if param == "" {
		return tag
	}
	return tag + "-" + param
}

// Errors validates a struct and returns structured field errors. Unlike Struct
// it never includes field values, so it is safe to serialise directly into an
// API response. A non-validation error is returned as a single FieldError whose
// Field is empty and whose Tag carries the message.
func Errors(data any) ([]FieldError, error) {
	err := validate.Struct(data)
	if err == nil {
		return nil, nil
	}

	var vErrs validator.ValidationErrors
	if !errors.As(err, &vErrs) {
		return nil, err
	}

	out := make([]FieldError, 0, len(vErrs))
	for _, e := range vErrs {
		out = append(out, FieldError{Field: e.Field(), Tag: e.Tag(), Param: e.Param()})
	}
	return out, nil
}

// Struct validates a struct's exported fields based on their 'validate' tags.
// If validation fails, it returns a formatted error string containing
// the field name, the actual value (redacted for sensitive fields), and the
// failed validation tag/parameter.
func Struct(data any) error {
	errs := validate.Struct(data)
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
		tag := e.Tag()
		b.WriteString("[")
		b.WriteString(field)

		if v := fmt.Sprint(e.Value()); v != "" {
			b.WriteString(":")
			b.WriteString(redact(field, v))
		}
		b.WriteString("]:[")
		b.WriteString(joinTag(tag, param))
		b.WriteString("]")
		results = append(results, b.String())
	}

	return errors.New(strings.Join(results, "\n"))
}

// MustStruct validates a struct's exported fields based on their 'validate' tags.
func MustStruct(data any) {
	err := Struct(data)
	if err != nil {
		panic(err)
	}
}

// Var validates a single variable against a specific validation tag.
// Example: Var("admin@example.com", "required,email")
func Var(s any, tag string) error {
	return validate.Var(s, tag)
}

// MustVar validates a single variable against a specific validation tag.
func MustVar(s any, tag string) {
	err := Var(s, tag)
	if err != nil {
		panic(err)
	}
}
