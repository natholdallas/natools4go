// Package narder provides a tiny leveled logger shared across the natools4go
// toolkits. It exposes a minimal surface (Info/Warn/Error with fmt variants)
// while routing output to the standard library logger. Call sites are resolved
// automatically: any frame inside the natools4go module is skipped, so the
// reported file:line always points at the user's code, even when logging is
// wrapped (e.g. via dbg).
package narder

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
)

const (
	// LevelInfo is the severity label used for informational messages.
	LevelInfo = "INFO"
	// LevelWarn is the severity label used for warning messages.
	LevelWarn = "WARN"
	// LevelError is the severity label used for error messages.
	LevelError = "ERROR"
)

// callerPredicate reports whether a frame belongs to this module's internals
// and should be skipped when resolving the call site. It is a variable so
// tests can replace it to exercise the machinery without depending on the
// module path. Frames from _test.go files are never skipped, so tests inside
// the module still resolve to their own line.
var callerPredicate = func(name, file string) bool {
	return strings.HasPrefix(name, "github.com/natholdallas/natools4go/") &&
		!strings.HasSuffix(file, "_test.go")
}

var (
	stdlib atomic.Pointer[log.Logger]
	printer atomic.Pointer[printerFunc]
	debug   atomic.Bool
)

type printerFunc func(level, file string, line int, msg string)

func init() {
	stdlib.Store(log.New(os.Stderr, "", log.Ldate|log.Ltime))
}

// defaultLogger returns a fresh stdlib logger writing to os.Stderr with the
// default flags (Ldate|Ltime, no call site).
func defaultLogger() *log.Logger {
	return log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

// SetOutput redirects the logger output, preserving the current flags.
// nil restores the default (os.Stderr).
func SetOutput(w io.Writer) {
	l := defaultLogger()
	if w != nil {
		l = log.New(w, "", log.Ldate|log.Ltime)
	}
	if cur := stdlib.Load(); cur != nil {
		l.SetFlags(cur.Flags())
	}
	stdlib.Store(l)
	printer.Store(nil)
}

// SetFlags sets the flags of the stdlib logger (e.g. log.Lshortfile to include
// the caller file:line). The reported position is the user's call site, not
// narder's or any wrapping package's internal frame.
func SetFlags(flags int) {
	l := stdlib.Load()
	l.SetFlags(flags)
	stdlib.Store(l)
}

// SetLogger replaces the stdlib backend. nil restores the default logger
// (stderr, Ldate|Ltime). A custom printer set via SetPrinter is cleared, since
// the printer and the stdlib backend are mutually exclusive.
func SetLogger(l *log.Logger) {
	if l == nil {
		l = defaultLogger()
	}
	stdlib.Store(l)
	printer.Store(nil)
}

// SetPrinter installs a custom sink instead of the stdlib logger. It receives
// the level, the resolved caller file:line, and the formatted message, letting
// users forward narder output to slog, zerolog, zap, or any structured logger.
// nil restores the stdlib backend.
func SetPrinter(p func(level, file string, line int, msg string)) {
	if p == nil {
		printer.Store(nil)
		return
	}
	printer.Store((*printerFunc)(&p))
}

// SetDebugMode toggles debug mode globally. Consumers such as fext use this
// to decide whether to surface extra diagnostic detail (e.g. internal system
// errors in responses). It is orthogonal to the log levels.
func SetDebugMode(mode bool) {
	debug.Store(mode)
}

// IsDebug reports whether debug mode is enabled.
func IsDebug() bool {
	return debug.Load()
}

// Info logs a message at INFO level.
func Info(v ...any) { emit(LevelInfo, fmt.Sprint(v...)) }

// Infof logs a formatted message at INFO level.
func Infof(format string, v ...any) { emit(LevelInfo, fmt.Sprintf(format, v...)) }

// Warn logs a message at WARN level.
func Warn(v ...any) { emit(LevelWarn, fmt.Sprint(v...)) }

// Warnf logs a formatted message at WARN level.
func Warnf(format string, v ...any) { emit(LevelWarn, fmt.Sprintf(format, v...)) }

// Error logs a message at ERROR level.
func Error(v ...any) { emit(LevelError, fmt.Sprint(v...)) }

// Errorf logs a formatted message at ERROR level.
func Errorf(format string, v ...any) { emit(LevelError, fmt.Sprintf(format, v...)) }

// emit resolves the caller position, skipping every frame inside the
// natools4go module, then dispatches to the active backend.
func emit(level, msg string) {
	file, line, depth := caller(0)
	if p := printer.Load(); p != nil {
		(*p)(level, file, line, msg)
		return
	}
	l := stdlib.Load()
	// runtime.Caller(0) inside caller() is caller itself, and inside Output it
	// is Output itself; both count the shared frames (emit, Info/Warn/Error)
	// identically, so depth maps 1:1 to Output's calldepth.
	_ = l.Output(depth, "["+level+"] "+msg)
}

// caller walks up the stack skipping module-internal frames and returns the
// first user frame's file, line, and its depth relative to start.
func caller(start int) (file string, line int, depth int) {
	for i := start; ; i++ {
		pc, f, l, ok := runtime.Caller(i)
		if !ok {
			return file, line, depth
		}
		if f != "" && l > 0 {
			file, line = f, l
		}
		depth = i - start
		fn := runtime.FuncForPC(pc)
		if fn == nil || !callerPredicate(fn.Name(), f) {
			return file, line, depth
		}
	}
}

