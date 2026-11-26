package flags

import (
	"errors"
	"testing"
)

func TestRunExecutesWhenTrueAndContinues(t *testing.T) {
	called := false
	Run(true, func() { called = true }, true) // continuer=true -> no os.Exit
	if !called {
		t.Error("script was not executed")
	}
}

func TestRunSkipsWhenFalse(t *testing.T) {
	called := false
	Run(false, func() { called = true })
	if called {
		t.Error("script should not execute when s is false")
	}
}

func TestRunESkipsWhenFalse(t *testing.T) {
	called := false
	err := RunE(false, func() error {
		called = true
		return nil
	}, true)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if called {
		t.Error("script should not execute when s is false")
	}
}

func TestRunEReturnsScriptError(t *testing.T) {
	want := errors.New("boom")
	got := RunE(true, func() error { return want }, true)
	if !errors.Is(got, want) {
		t.Errorf("err = %v, want %v", got, want)
	}
}
