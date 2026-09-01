// Package dbg provides high-visibility debugging and file inspection tools.
// All output is routed through the shared narder logger; call sites are
// resolved automatically to the caller of these functions.
package dbg

import (
	"bufio"
	"io"
	"os"

	"github.com/natholdallas/natools4go/jsons"
	"github.com/natholdallas/natools4go/narder"
)

// Err prints each non-nil error from the provided slice at ERROR level.
func Err(errs ...error) {
	for _, err := range errs {
		if err != nil {
			narder.Error(err)
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
			narder.Errorf("[JSON-FAIL] %v", err)
			continue
		}
		narder.Info(d)
	}
}

// Struct prints the structural representation of values using the %+v format,
// which includes field names for structs.
func Struct(v ...any) {
	for _, i := range v {
		narder.Infof("%+v", i)
	}
}

// File reads the file at the specified path and prints its content, behaving
// similarly to the Unix 'cat' command.
func File(path string) {
	file, err := os.Open(path)
	if err != nil {
		narder.Errorf("[FILE-ERR] %s: %v", path, err)
		return // Crucial: stop execution to avoid nil pointer panic
	}
	defer file.Close()

	// bufio.Reader has no line-length limit (bufio.Scanner truncates at 64KB),
	// so arbitrarily long lines are printed faithfully.
	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			narder.Info(line)
		}
		if err != nil {
			if err != io.EOF {
				narder.Errorf("[SCAN-ERR] %s: %v", path, err)
			}
			return
		}
	}
}