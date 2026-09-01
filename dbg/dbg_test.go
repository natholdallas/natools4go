package dbg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natholdallas/natools4go/narder"
)

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

var out bytes.Buffer

// capture redirects narder output into a buffer for assertions.
func capture() {
	out.Reset()
	narder.SetOutput(&out)
}

func readOut() string {
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

func TestFileLongLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "long.txt")
	// >64KB line: bufio.Scanner would truncate this; the fix must not.
	long := strings.Repeat("x", 70*1024) + "\n"
	if err := os.WriteFile(path, []byte(long), 0o644); err != nil {
		t.Fatal(err)
	}
	capture()
	File(path)
	got := readOut()
	if !strings.Contains(got, long) {
		t.Errorf("long line truncated: got %d bytes, want to contain %d", len(got), len(long))
	}
}

func TestDump(t *testing.T) {
	capture()
	Dump(42)
	if got := readOut(); got == "" {
		t.Error("Dump should produce output")
	}
}