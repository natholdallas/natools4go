package ask

import (
	"os"
	"strings"
	"testing"
)

// withStdin replaces os.Stdin with a pipe fed by input for the duration of fn.
func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Close()

	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	fn()
}

func TestReadString(t *testing.T) {
	withStdin(t, "hello\n", func() {
		if got := Read[string]("name"); got != "hello" {
			t.Errorf("Read = %q, want hello", got)
		}
	})
}

func TestReadInt(t *testing.T) {
	withStdin(t, "42\n", func() {
		if got := Read[int]("age"); got != 42 {
			t.Errorf("Read = %d, want 42", got)
		}
	})
}

func TestLine(t *testing.T) {
	withStdin(t, "  full line of text  \n", func() {
		if got := Line("msg"); got != "full line of text" {
			t.Errorf("Line = %q, want trimmed text", got)
		}
	})
}

func TestConfirm(t *testing.T) {
	withStdin(t, "y\n", func() {
		if !Confirm("continue?") {
			t.Error("Confirm should be true for y")
		}
	})
	withStdin(t, "no\n", func() {
		if Confirm("continue?") {
			t.Error("Confirm should be false for no")
		}
	})
	withStdin(t, "Y\n", func() {
		if !Confirm("continue?") {
			t.Error("Confirm should be true for uppercase Y")
		}
	})
}

func TestLineEmpty(t *testing.T) {
	withStdin(t, "\n", func() {
		if got := Line("msg"); got != "" {
			t.Errorf("Line = %q, want empty", got)
		}
	})
}

var _ = strings.TrimSpace // keep import if used later