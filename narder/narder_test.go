package narder

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func captureOut(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	SetOutput(buf)
	t.Cleanup(func() { SetOutput(os.Stderr) })
	return buf
}

func TestLevelsAndPrefixes(t *testing.T) {
	buf := captureOut(t)
	Info("hello", "world")
	Warn("careful")
	Error("boom")
	got := buf.String()
	for _, want := range []string{"[INFO] helloworld", "[WARN] careful", "[ERROR] boom"} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want to contain %q", got, want)
		}
	}
}

func TestFormatVariants(t *testing.T) {
	buf := captureOut(t)
	Infof("a=%d", 1)
	Warnf("b=%s", "x")
	Errorf("c=%t", true)
	got := buf.String()
	for _, want := range []string{"[INFO] a=1", "[WARN] b=x", "[ERROR] c=true"} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want to contain %q", got, want)
		}
	}
}

func TestDefaultFlagsNoCallSite(t *testing.T) {
	buf := captureOut(t)
	Info("x")
	if strings.Contains(buf.String(), "narder_test.go") {
		t.Errorf("default flags should not print file:line, got %q", buf.String())
	}
}

func TestFlagsPrintsUserCallSite(t *testing.T) {
	buf := captureOut(t)
	SetFlags(log.Lshortfile)
	t.Cleanup(func() { SetFlags(0) })
	_, _, line, _ := runtime.Caller(0)
	Info("x") // call site is the line right below
	got := buf.String()
	want := "narder_test.go:" + strconv.Itoa(line+1)
	if !strings.Contains(got, want) {
		t.Errorf("expected call site %q in output, got %q", want, got)
	}
}

func TestCallerResolvesTestFrame(t *testing.T) {
	file, _, depth := caller(0)
	if depth != 1 {
		t.Errorf("depth = %d, want 1 (caller itself, then the test frame)", depth)
	}
	if !strings.HasSuffix(file, "narder_test.go") {
		t.Errorf("caller should resolve to test file, got %q", file)
	}
}

func TestPrinterHook(t *testing.T) {
	var got []string
	SetPrinter(func(level, file string, line int, msg string) {
		got = append(got, level+"|"+filepath.Base(file)+"|"+strconv.Itoa(line)+"|"+msg)
	})
	t.Cleanup(func() {
		SetPrinter(nil)
		SetOutput(os.Stderr)
	})
	_, _, line, _ := runtime.Caller(0)
	Info("v")
	Errorf("e=%d", 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 printer calls, got %d: %v", len(got), got)
	}
	if want := "INFO|narder_test.go|" + strconv.Itoa(line+1) + "|v"; got[0] != want {
		t.Errorf("printer got %q, want %q", got[0], want)
	}
	if !strings.HasPrefix(got[1], "ERROR|narder_test.go|") || !strings.HasSuffix(got[1], "|e=2") {
		t.Errorf("printer got %q, want ERROR with file and e=2", got[1])
	}
}

func TestSetLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	SetLogger(log.New(buf, "PREFIX ", 0))
	t.Cleanup(func() { SetLogger(nil) })
	Info("x")
	if !strings.Contains(buf.String(), "PREFIX [INFO] x") {
		t.Errorf("SetLogger output = %q, want PREFIX [INFO] x", buf.String())
	}
}

func TestSetOutputNilRestoresDefault(t *testing.T) {
	SetOutput(nil) // must not panic
	if l := stdlib.Load(); l == nil {
		t.Fatal("stdlib logger should never be nil")
	}
	if l := printer.Load(); l != nil {
		t.Fatal("printer should be cleared by SetOutput")
	}
}

func TestDebugMode(t *testing.T) {
	SetDebugMode(false)
	if IsDebug() {
		t.Error("IsDebug should be false after SetDebugMode(false)")
	}
	SetDebugMode(true)
	if !IsDebug() {
		t.Error("IsDebug should be true after SetDebugMode(true)")
	}
	SetDebugMode(false) // restore default
}

func TestCallerPredicate(t *testing.T) {
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"github.com/natholdallas/natools4go/dbg.JSON", "/go/natools4go/dbg/dbg.go", true},
		{"github.com/natholdallas/natools4go/fext.Listen", "/go/natools4go/fext/ctx.go", true},
		{"github.com/natholdallas/natools4go/narder.emit", "/go/natools4go/narder/narder.go", true},
		{"github.com/natholdallas/natools4go/narder.TestX", "/go/natools4go/narder/narder_test.go", false},
		{"main.main", "/home/app/main.go", false},
		{"github.com/other/proj/util.Log", "/src/other/proj/util.go", false},
	}
	for _, c := range cases {
		if got := callerPredicate(c.name, c.file); got != c.want {
			t.Errorf("callerPredicate(%q, %q) = %v, want %v", c.name, c.file, got, c.want)
		}
	}
}
