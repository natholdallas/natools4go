// Package spew provides high-visibility debugging and file inspection tools.
package spew

import (
	"bufio"
	"fmt"
	"os"

	"github.com/natholdallas/natools4go/jsons"
)

// Err prints each non-nil error from the provided slice to standard output.
func Err(errs ...error) {
	for _, err := range errs {
		if err != nil {
			// Using Fprintln to os.Stderr separates errors from normal output
			fmt.Fprintln(os.Stderr, "[ERROR]", err)
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
			// Fallback to Go syntax representation if JSON fails
			fmt.Printf("[JSON-FAIL] %#v\n", i)
			continue
		}
		fmt.Println(d)
	}
}

// Struct prints the structural representation of values using the %+v format,
// which includes field names for structs.
func Struct(v ...any) {
	for _, i := range v {
		fmt.Printf("%+v\n", i)
	}
}

// File reads the file at the specified path and prints its content to standard output,
// behaving similarly to the Unix 'cat' command.
func File(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FILE-ERR] %s: %v\n", path, err)
		return // Crucial: stop execution to avoid nil pointer panic
	}
	defer file.Close()

	// Using a dedicated buffer can be faster for large files
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[SCAN-ERR] %s: %v\n", path, err)
	}
}
