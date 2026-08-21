package spew

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

var (
	outMu sync.Mutex
	out   strings.Builder
)

// capture routes the printer into a buffer for assertions.
func capture() {
	outMu.Lock()
	out.Reset()
	outMu.Unlock()
	SetPrinter(func(format string, v ...any) {
		outMu.Lock()
		defer outMu.Unlock()
		out.WriteString(fmt.Sprintf(format, v...))
	})
}

func readOut() string {
	outMu.Lock()
	defer outMu.Unlock()
	return out.String()
}

func TestErrPrintsOnlyNonNil(t *testing.T) {
	capture()
	Err(nil, &testError{msg: "boom"}, nil)
	got := readOut()
	if !strings.Contains(got, "[ERROR] boom") {
		t.Errorf("Err output = %q, want [ERROR] boom", got)
	}
	if strings.Count(got, "[ERROR]") != 1 {
		t.Errorf("expected exactly 1 error line, got %q", got)
	}
}

func TestStructPrintsFields(t *testing.T) {
	capture()
	Struct(struct{ A, B int }{1, 2})
	got := readOut()
	if !strings.Contains(got, "A:1") || !strings.Contains(got, "B:2") {
		t.Errorf("Struct output = %q, want fields A:1 B:2", got)
	}
}

func TestJSONPrintsJSON(t *testing.T) {
	capture()
	JSON(map[string]int{"x": 1})
	got := readOut()
	if !strings.Contains(got, `"x"`) || !strings.Contains(got, "1") {
		t.Errorf("JSON output = %q, want JSON with x=1", got)
	}
}

func TestFilePrintsContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	capture()
	File(path)
	got := readOut()
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") {
		t.Errorf("File output = %q, want line1 and line2", got)
	}
}

func TestFileMissing(t *testing.T) {
	capture()
	File("/nonexistent/definitely-missing.txt")
	got := readOut()
	if !strings.Contains(got, "[FILE-ERR]") {
		t.Errorf("missing file should print [FILE-ERR], got %q", got)
	}
}

func TestDump(t *testing.T) {
	capture()
	Dump(42)
	if got := readOut(); got == "" {
		t.Error("Dump should produce output")
	}
}