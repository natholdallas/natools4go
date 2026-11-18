package ask

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
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

// ReadPassword prompts for a password without echoing it to the terminal.
//
// Use it instead of Read[string] whenever the value is a credential: Read uses
// fmt.Scanln, which prints what the user types (and leaves it in shell history
// and terminal scrollback).
//
// When stdin is not a terminal (piped input, CI) it falls back to reading a
// line, because terminal mode is unavailable there.
func ReadPassword(field string) (string, error) {
	fmt.Printf("%s: ", field)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Not a TTY: read a plain line so `echo pw | ./backend --adm` still works.
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	raw, err := term.ReadPassword(fd)
	fmt.Println() // ReadPassword consumes the newline without echoing it.
	if err != nil {
		return "", err
	}
	return string(raw), nil
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
