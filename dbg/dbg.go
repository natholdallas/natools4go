// Package dbg provides high-visibility debugging and file inspection tools.
package dbg

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/natholdallas/natools4go/jsons"
)

var printer func(format string, v ...any) = func(format string, v ...any) { fmt.Printf(format, v...) }

// SetPrinter overrides the default output printer. It is intended to be
// called once during initialization, before any concurrent printing begins;
// concurrent SetPrinter/print is unsupported (the printer is read lock-free
// on the printing hot path).
func SetPrinter(p func(format string, v ...any)) {
	printer = p
}

// Err prints each non-nil error from the provided slice to standard output.
func Err(errs ...error) {
	for _, err := range errs {
		if err != nil {
			printer("[ERROR] %v\n", err)
		}
	}
}

// Dump is a convenience wrapper that calls JSON to print a detailed
// representation of the provided values.
func Dump(v ...any) {
	JSON(v...)
}

// JSON serializes each value into a formatted JSON string and prints it.
// Note: Only exported (public) fields will be included in the output.
func JSON(v ...any) {
	for _, i := range v {
		d, err := jsons.String(i, true)
		if err != nil {
			printer("[JSON-FAIL] %v\n", err)
			continue
		}
		printer("%s\n", d)
	}
}

// Struct prints the structural representation of values using the %+v format,
// which includes field names for structs.
func Struct(v ...any) {
	for _, i := range v {
		printer("%+v\n", i)
	}
}

// File reads the file at the specified path and prints its content to standard output,
// behaving similarly to the Unix 'cat' command.
func File(path string) {
	file, err := os.Open(path)
	if err != nil {
		printer("[FILE-ERR] %s: %v\n", path, err)
		return // Crucial: stop execution to avoid nil pointer panic
	}
	defer file.Close()

	// bufio.Reader has no line-length limit (bufio.Scanner truncates at 64KB),
	// so arbitrarily long lines are printed faithfully.
	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			printer("%s", line)
		}
		if err != nil {
			if err != io.EOF {
				printer("[SCAN-ERR] %s: %v\n", path, err)
			}
			return
		}
	}
}
