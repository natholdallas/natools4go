package flags

import "testing"

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