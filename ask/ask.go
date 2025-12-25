// Package ask is tiny packaging support fmt
package ask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Read prompts the user with a field name and scans a single value into type T.
// Note: This uses fmt.Scanln which may leave remaining data in the buffer
// if multiple words are entered.
func Read[T any](field string) T {
	var v T
	fmt.Print(field + ": ")
	fmt.Scanln(&v)
	return v
}

// Line prompts the user and reads an entire line of text as a string.
// This is more robust than Read[string] for inputs containing spaces.
func Line(field string) string {
	fmt.Printf("%s: ", field)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// Confirm prompts the user for a yes/no response.
// It returns true for "y", "Y", "yes", or "Yes".
func Confirm(message string) bool {
	fmt.Printf("%s (y/n): ", message)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
