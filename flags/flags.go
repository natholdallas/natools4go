// Package flags provides a set of functions to parse command line arguments
package flags

import (
	"fmt"
	"os"
)

func Run(s bool, script func(), continuer ...bool) {
	if !s {
		return
	}
	script()
	c := false
	if len(continuer) > 0 {
		c = continuer[0]
	}
	if !c {
		os.Exit(0)
	}
}

// RunE is Run for scripts that can fail. It exits 0 on success and prints the
// error to stderr and exits 1 on failure, so a failed migration or seed is
// never reported as success. Pass continuer=true to return the error instead
// of exiting (used by tests).
func RunE(s bool, script func() error, continuer ...bool) error {
	if !s {
		return nil
	}
	err := script()
	if len(continuer) > 0 && continuer[0] {
		return err
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
	return nil
}
