package va

import (
	"strings"
	"testing"
)

type creds struct {
	Username string `validate:"required"`
	Password string `validate:"required,min=4,max=20"`
}

// TestStructRedactsSensitiveValues guards against leaking credentials: the
// formatted error used to embed the offending value verbatim, and the project
// returns that string straight to HTTP clients.
func TestStructRedactsSensitiveValues(t *testing.T) {
	err := Struct(creds{Username: "bob", Password: "abc"})
	if err == nil {
		t.Fatal("expected a validation error")
	}
	msg := err.Error()
	if strings.Contains(msg, "abc") {
		t.Errorf("password leaked into validation error: %s", msg)
	}
	if !strings.Contains(msg, "***") {
		t.Errorf("expected redaction marker in %q", msg)
	}
	if !strings.Contains(msg, "Password") || !strings.Contains(msg, "min-4") {
		t.Errorf("error should still name the field and rule: %s", msg)
	}
}

// TestStructKeepsNormalValues ensures redaction is targeted and diagnostics for
// ordinary fields remain useful.
func TestStructKeepsNormalValues(t *testing.T) {
	type field struct {
		Username string `validate:"min=10"`
	}
	err := Struct(field{Username: "bob"})
	if err == nil {
		t.Fatal("expected a validation error")
	}
	if !strings.Contains(err.Error(), "bob") {
		t.Errorf("non-sensitive value should be shown: %s", err.Error())
	}
}

// TestErrorsReturnsStructuredOutput covers the value-free API meant for JSON
// responses.
func TestErrorsReturnsStructuredOutput(t *testing.T) {
	errs, err := Errors(creds{Username: "bob", Password: "abc"})
	if err != nil {
		t.Fatalf("Errors: %v", err)
	}
	if len(errs) != 1 {
		t.Fatalf("got %d field errors, want 1", len(errs))
	}
	got := errs[0]
	if got.Field != "Password" || got.Tag != "min" || got.Param != "4" {
		t.Errorf("FieldError = %+v, want {Password min 4}", got)
	}
	if strings.Contains(got.String(), "abc") {
		t.Errorf("FieldError.String leaked the value: %s", got.String())
	}
}

func TestErrorsNilWhenValid(t *testing.T) {
	errs, err := Errors(creds{Username: "bob", Password: "longenough"})
	if err != nil {
		t.Fatalf("Errors: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no field errors, got %+v", errs)
	}
}

// TestStructRedactionFieldNames pins the field-name heuristic in both
// directions: "Passport" must keep its value (the old regex matched its
// "pass" prefix), while "Pwd" must be redacted (the old regex missed it).
func TestStructRedactionFieldNames(t *testing.T) {
	type fields struct {
		Passport string `validate:"min=10"`
		Pwd      string `validate:"min=10"`
	}
	err := Struct(fields{Passport: "ab", Pwd: "cd"})
	if err == nil {
		t.Fatal("expected validation errors")
	}
	msg := err.Error()
	if !strings.Contains(msg, "[Passport:ab]") {
		t.Errorf("Passport value should be preserved: %s", msg)
	}
	if strings.Contains(msg, "cd") {
		t.Errorf("Pwd value leaked into validation error: %s", msg)
	}
	if !strings.Contains(msg, "[Pwd:***]") {
		t.Errorf("Pwd value should be redacted: %s", msg)
	}
}
