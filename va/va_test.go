package va

import (
	"strings"
	"testing"
)

type signUp struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18"`
}

func TestStructValid(t *testing.T) {
	if err := Struct(signUp{Email: "a@b.com", Age: 20}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStructInvalid(t *testing.T) {
	err := Struct(signUp{Email: "not-an-email", Age: 10})
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "[Email") || !strings.Contains(msg, "email") {
		t.Errorf("error should mention Email/email, got %q", msg)
	}
	if !strings.Contains(msg, "[Age") || !strings.Contains(msg, "gte-18") {
		t.Errorf("error should mention Age/gte-18, got %q", msg)
	}
}

func TestStructNilData(t *testing.T) {
	if err := Struct((*signUp)(nil)); err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

func TestVar(t *testing.T) {
	if err := Var("admin@example.com", "required,email"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := Var("bad", "email"); err == nil {
		t.Error("expected error for invalid email")
	}
}